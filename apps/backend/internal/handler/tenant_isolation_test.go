//nolint:lll // Explicit test-only scoped SQL and HTTP matrix preserve the reviewed wire requests.
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/database"
	"github.com/6sLOGAN78/flux/internal/repository"
	fluxTesting "github.com/6sLOGAN78/flux/internal/testing"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

// TestControlPlaneTenantDenials composes production HTTP authorization over
// independently migrated PostgreSQL, with two disjoint durable identities.
func TestControlPlaneTenantDenials(t *testing.T) {
	db, cleanup := fluxTesting.SetupTestDB(t)
	defer cleanup()
	p := newSignedProvider(t)
	db.Config.Server.CORSAllowedOrigins = []string{fixtureParty}
	api := httptest.NewServer(productRouter(db.Config, &database.Database{Pool: db.Pool}, p))
	defer api.Close()
	ctx := context.Background()
	tenants := seedDenialTenants(t, db, api.URL, p)
	t.Run("missing-scope-never-defaults-to-global", func(t *testing.T) {
		store := repository.NewWorkspaceRepository(db.Pool)
		for _, scope := range []repository.Scope{
			{}, {WorkspaceID: uuid.MustParse(tenants[0].workspace)},
			{ActorID: uuid.MustParse(tenants[0].actor)},
		} {
			tx, err := db.Pool.Begin(ctx)
			require.NoError(t, err)
			result, scopeErr := store.LockScope(ctx, tx, scope, false)
			require.ErrorIs(t, scopeErr, pgx.ErrNoRows)
			require.Equal(t, uuid.Nil, result.ID)
			require.NoError(t, tx.Rollback(ctx))
		}
	})
	for i := range tenants {
		own, foreign := tenants[i], tenants[1-i]
		t.Run(fmt.Sprintf("tenant-%d-direct-IDs-and-private-errors", i), func(t *testing.T) {
			ownStatus, _ := workspaceRequest(t, api.URL, own.token, "PUT", "/me/last-workspace", "", `{"workspaceId":"`+own.workspace+`"}`)
			require.Equal(t, 200, ownStatus)
			for n, tc := range []struct {
				method, path, body string
				status             int
			}{
				{"GET", foreign.path(), "", 404},
				{"PUT", "/me/last-workspace", `{"workspaceId":"` + foreign.workspace + `"}`, 404},
				{"GET", foreign.path() + "/links", "", 404},
				{"GET", own.path() + "/links/" + foreign.link, "", 404},
				{"POST", foreign.path() + "/links", own.createBody, 404},
				{"POST", own.path() + "/links/" + foreign.link + "/transitions", `{"action":"disable","expectedVersion":"1"}`, 404},
				{"POST", own.path() + "/links/" + foreign.link + "/suspension", `{"reason":"Private denial reason","expectedVersion":"1"}`, 404},
				{"DELETE", own.path() + "/links/" + foreign.link + "/suspension", `{"reason":"Private denial reason","expectedVersion":"1"}`, 404},
				{"GET", foreign.path() + "/members", "", 404},
				{"PATCH", own.path() + "/members/" + foreign.member, `{"role":"member"}`, 404},
				{"DELETE", own.path() + "/members/" + foreign.member, "{}", 404},
				{"GET", foreign.path() + "/invitations", "", 404},
				{"POST", foreign.path() + "/invitations", `{"email":"private@example.test","role":"member"}`, 404},
				{"POST", own.path() + "/invitations/" + foreign.invitation + "/resend", `{"version":1}`, 404},
				{"DELETE", own.path() + "/invitations/" + foreign.invitation, `{"version":1}`, 404},
			} {
				code, body := workspaceRequest(t, api.URL, own.token, tc.method, tc.path, uuid.NewString(), tc.body)
				require.Equal(t, tc.status, code, "denial matrix case %d", n)
				encoded, err := json.Marshal(body)
				require.NoError(t, err)
				assertDenialPrivate(t, string(encoded), foreign)
			}
		})
		t.Run(fmt.Sprintf("tenant-%d-search-cursor-and-ledger", i), func(t *testing.T) {
			code, response := workspaceRequest(t, api.URL, own.token, "GET", own.path()+"/links?limit=1", "", "")
			require.Equal(t, 200, code)
			cursor := response["nextCursor"].(string)
			// The caller is authorized in this second workspace: cursor scope itself
			// must reject transplantation, independently of the membership gate.
			_, err := db.Pool.Exec(ctx, "INSERT INTO memberships(workspace_id,user_id,role,created_by) VALUES($1,$2,'viewer',$3)", foreign.workspace, own.actor, foreign.actor)
			require.NoError(t, err)
			code, response = workspaceRequest(t, api.URL, own.token, "GET", foreign.path()+"/links?limit=1&cursor="+url.QueryEscape(cursor), "", "")
			require.Equal(t, 400, code)
			require.Equal(t, "CURSOR_INVALID", response["code"])
			_, err = db.Pool.Exec(ctx, "DELETE FROM memberships WHERE workspace_id=$1 AND user_id=$2", foreign.workspace, own.actor)
			require.NoError(t, err)
			code, response = workspaceRequest(t, api.URL, own.token, "GET", own.path()+"/links?search="+url.QueryEscape(fmt.Sprintf("denial-%d-", 1-i)), "", "")
			require.Equal(t, 200, code)
			require.Empty(t, response["items"])
			code, response = workspaceRequest(t, api.URL, own.token, "POST", foreign.path()+"/links", foreign.createKey, foreign.createBody)
			require.Equal(t, 404, code)
			encoded, err := json.Marshal(response)
			require.NoError(t, err)
			assertDenialPrivate(t, string(encoded), foreign)
			code, response = workspaceRequest(t, api.URL, own.token, "POST", own.path()+"/links", foreign.createKey, own.createBody)
			require.Equal(t, 201, code, "another tenant's retry key must not select its ledger")
			require.NotEqual(t, foreign.link, response["link"].(map[string]any)["id"])
			first := response
			code, response = workspaceRequest(t, api.URL, own.token, "POST", own.path()+"/links", foreign.createKey, own.createBody)
			require.Equal(t, 201, code)
			require.Equal(t, first, response, "replay belongs to the authorized workspace")
		})
	}
	t.Run("composite-FKs-and-owned-row-workspaces", func(t *testing.T) {
		for _, statement := range []string{
			"INSERT INTO audit_events(workspace_id,actor_id,operation,target_link_id) VALUES($1,$2,'link.disable',$3)",
			"INSERT INTO audit_events(workspace_id,actor_id,operation,target_invitation_id) VALUES($1,$2,'invitation.revoke',$3)",
			"INSERT INTO invitation_delivery_intents(workspace_id,id,invitation_id,key_id,ciphertext) SELECT $1,gen_random_uuid(),$3,'fixture',decode(repeat('00',28),'hex') WHERE $2::uuid IS NOT NULL",
		} {
			target := tenants[1].link
			if strings.Contains(statement, "invitation") {
				target = tenants[1].invitation
			}
			_, err := db.Pool.Exec(ctx, statement, tenants[0].workspace, tenants[0].actor, target)
			var constraint *pgconn.PgError
			require.ErrorAs(t, err, &constraint)
			require.Equal(t, "23503", constraint.Code)
		}
		for _, table := range []string{"memberships", "invitations", "invitation_delivery_intents", "links", "mutation_requests", "audit_events", "workspace_bootstrap_requests"} {
			var rows, unscoped int
			require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*),count(*) FILTER(WHERE workspace_id IS NULL) FROM "+table).Scan(&rows, &unscoped))
			require.Positive(t, rows, "owned fixture table %s must actually contain records", table)
			require.Zero(t, unscoped)
		}
	})
	t.Run("locked-current-actor-removal-beats-HTTP-mutation", func(t *testing.T) {
		checkHTTPRemovalBarrier(t, db, api.URL, tenants[0])
	})
	t.Run("absent-future-surfaces-and-operational-log-privacy", func(t *testing.T) {
		own := tenants[1]
		for _, path := range []string{own.path() + "/analytics", own.path() + "/cache", "/redirects", "/clicks", "/conversions"} {
			code, _ := workspaceRequest(t, api.URL, own.token, "GET", path, "", "")
			require.Equal(t, 404, code)
		}
		require.NotEmpty(t, p.logs.String(), "operational logs must actually be captured")
		for _, tenant := range tenants {
			assertDenialPrivate(t, p.logs.String(), tenant)
		}
	})
}

type denialTenant struct {
	token, workspace, actor, member, link, invitation, createKey, createBody string
}

func (tenant denialTenant) path() string { return "/workspaces/" + tenant.workspace }
func assertDenialPrivate(t *testing.T, surface string, tenant denialTenant) {
	t.Helper()
	for _, value := range []string{tenant.token, tenant.workspace, tenant.actor, tenant.member, tenant.link, tenant.invitation, "Private denial", "denial-0@example.test", "denial-1@example.test", "https://example.com/denial"} {
		leaked := strings.Contains(surface, value)
		require.False(t, leaked, "private surface must not disclose a tenant field")
	}
}

func checkHTTPRemovalBarrier(t *testing.T, db *fluxTesting.TestDB, api string, own denialTenant) {
	t.Helper()
	ctx := context.Background()
	lock, err := db.Pool.Begin(ctx)
	require.NoError(t, err)
	defer func() { _ = lock.Rollback(ctx) }()
	_, err = lock.Exec(ctx, "SELECT id FROM workspaces WHERE id=$1 FOR UPDATE", own.workspace)
	require.NoError(t, err)
	done := make(chan int, 1)
	key := uuid.NewString()
	go func() {
		req, requestErr := http.NewRequestWithContext(ctx, http.MethodPost, api+"/api/v1"+own.path()+"/links/"+own.link+"/transitions", bytes.NewBufferString(`{"action":"disable","expectedVersion":"1"}`))
		if requestErr != nil {
			done <- 0
			return
		}
		req.Header.Set("Authorization", "Bearer "+own.token)
		req.Header.Set("Origin", fixtureParty)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", key)
		response, requestErr := (&http.Client{Timeout: 4 * time.Second}).Do(req)
		if requestErr != nil {
			done <- 0
			return
		}
		_ = response.Body.Close()
		done <- response.StatusCode
	}()
	// Use a valid lifecycle body; this separate helper sends only native HTTP.
	// The barrier below observes PostgreSQL lock waiting, never a timing guess.
	require.Eventually(t, func() bool {
		var waiting int
		return db.Pool.QueryRow(ctx, "SELECT count(*) FROM pg_stat_activity WHERE datname=current_database() AND wait_event_type='Lock'").Scan(&waiting) == nil && waiting > 0
	}, time.Second, 10*time.Millisecond)
	_, err = lock.Exec(ctx, "DELETE FROM memberships WHERE workspace_id=$1 AND user_id=$2", own.workspace, own.actor)
	require.NoError(t, err)
	require.NoError(t, lock.Commit(ctx))
	require.Equal(t, 404, <-done)
	var effects int
	require.NoError(t, db.Pool.QueryRow(ctx, "SELECT count(*) FROM mutation_requests WHERE workspace_id=$1 AND request_key=$2", own.workspace, key).Scan(&effects))
	require.Zero(t, effects)
	code, _ := workspaceRequest(t, api, own.token, "POST", own.path()+"/links", own.createKey, own.createBody)
	require.Equal(t, 404, code, "removed actor cannot read an existing ledger")
}

func seedDenialTenants(t *testing.T, db *fluxTesting.TestDB, api string, p *signedProvider) [2]denialTenant {
	t.Helper()
	ctx := context.Background()
	var tenants [2]denialTenant
	for i := range tenants {
		token := p.token(t, map[string]any{"sub": fmt.Sprintf("user_denial_%d", i), "sid": fmt.Sprintf("sess_user_denial_%d", i)})
		code, response := workspaceRequest(t, api, token, "POST", "/workspaces", uuid.NewString(), fmt.Sprintf(`{"name":"Private denial workspace %d"}`, i))
		require.Equal(t, 201, code)
		tenants[i] = denialTenant{token: token, workspace: response["workspace"].(map[string]any)["id"].(string)}
		require.NoError(t, db.Pool.QueryRow(ctx, "SELECT created_by FROM workspaces WHERE id=$1", tenants[i].workspace).Scan(&tenants[i].actor))
		for n := range 2 {
			body := fmt.Sprintf(`{"destination":"https://example.com/denial-%d-%d","title":"Private denial title %d"}`, i, n, i)
			key := uuid.NewString()
			code, response = workspaceRequest(t, api, token, "POST", tenants[i].path()+"/links", key, body)
			require.Equal(t, 201, code)
			if n == 0 {
				tenants[i].link = response["link"].(map[string]any)["id"].(string)
				tenants[i].createKey, tenants[i].createBody = key, body
			}
		}
		code, response = workspaceRequest(t, api, token, "POST", tenants[i].path()+"/invitations", uuid.NewString(), fmt.Sprintf(`{"email":"denial-%d@example.test","role":"member"}`, i))
		require.Equal(t, 201, code)
		tenants[i].invitation = response["invitation"].(map[string]any)["id"].(string)
		require.NoError(t, db.Pool.QueryRow(ctx, "SELECT id FROM memberships WHERE workspace_id=$1 AND user_id=$2", tenants[i].workspace, tenants[i].actor).Scan(&tenants[i].member))
	}
	return tenants
}
