//nolint:cyclop // The contract proof includes recursive schema validation and complete generator failure scenarios.
package transport_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/6sLOGAN78/flux/internal/transport"
	"gopkg.in/yaml.v3"
)

func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err27 := json.Unmarshal(data, &result); err27 != nil {
		t.Fatal(err27)
	}
	return result
}

// Check the schema from the canonical document, rather than a second health schema.
// Fail closed if this small fixture validator encounters an unsupported schema type.
//
//nolint:gocognit // Keep this regression scenario and its ordered failure assertions together.
func validateSchema(schema map[string]any, value any) error {
	switch schema["type"] {
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
		if values, ok := schema["enum"].([]any); ok {
			for _, allowed := range values {
				if allowed == value {
					return nil
				}
			}
			return errors.New("boolean is outside canonical enum")
		}
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
		if values, ok := schema["enum"].([]any); ok {
			for _, allowed := range values {
				if allowed == value {
					return nil
				}
			}
			return fmt.Errorf("value %v is outside enum %v", value, values)
		}
	case "object":
		object, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
		properties := schema["properties"].(map[string]any)
		for _, required := range schema["required"].([]any) {
			if _, exists := object[required.(string)]; !exists {
				return fmt.Errorf("missing required property %v", required)
			}
		}
		for name, field := range object {
			property, exists := properties[name]
			if !exists {
				if schema["additionalProperties"] == false {
					return fmt.Errorf("undocumented property %s", name)
				}
				continue
			}
			if err := validateSchema(property.(map[string]any), field); err != nil {
				return fmt.Errorf("%s: %w", name, err)
			}
		}
	case "array":
		items, ok := value.([]any)
		if !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
		for i, item := range items {
			if err := validateSchema(schema["items"].(map[string]any), item); err != nil {
				return fmt.Errorf("item %d: %w", i, err)
			}
		}
	default:
		return fmt.Errorf("unsupported canonical schema type %v", schema["type"])
	}
	return nil
}

func responseSchema(t *testing.T, document map[string]any, path, status, name string) map[string]any {
	t.Helper()
	paths := document["paths"].(map[string]any)
	operation := paths[path].(map[string]any)["get"].(map[string]any)
	response := operation["responses"].(map[string]any)[status].(map[string]any)
	content := response["content"].(map[string]any)["application/json"].(map[string]any)
	ref := content["schema"].(map[string]any)["$ref"]
	if want := "#/components/schemas/transport." + name; ref != want {
		t.Fatalf("%s HTTP %s schema = %v, want %s", path, status, ref, want)
	}
	components := document["components"].(map[string]any)
	return components["schemas"].(map[string]any)["transport."+name].(map[string]any)
}

//nolint:gocognit // Keep this regression scenario and its ordered failure assertions together.
func TestHealthJSONMatchesCanonicalResponses(t *testing.T) {
	document := readJSON(t, "../../../../packages/openapi/openapi.json")
	tests := []struct {
		value   any
		name    string
		path    string
		status  string
		schema  string
		fixture string
	}{
		{name: "live",
			path:    "/live",
			status:  "200",
			schema:  "HealthLiveResponse",
			fixture: `{"status":"alive"}`,
			value:   &transport.HealthLiveResponse{Status: "alive"}},
		{name: "ready",
			path:    "/ready",
			status:  "200",
			schema:  "HealthReadyResponse",
			fixture: `{"status":"ready","checks":[{"name":"database","state":"ready"}]}`,
			value:   &transport.HealthReadyResponse{}},
		{name: "not_ready",
			path:   "/ready",
			status: "503",
			schema: "HealthReadyResponse",
			fixture: `{"status":"not_ready","checks":[{"name":"database","state":"not_ready"},` +
				`{"name":"redis","state":"ready"}]}`,
			value: &transport.HealthReadyResponse{}},
		{name: "no_dependencies",
			path:    "/ready",
			status:  "200",
			schema:  "HealthReadyResponse",
			fixture: `{"status":"ready","checks":[]}`,
			value:   &transport.HealthReadyResponse{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := responseSchema(t, document, tt.path, tt.status, tt.schema)
			var fixture any
			if err := json.Unmarshal([]byte(tt.fixture), &fixture); err != nil {
				t.Fatal(err)
			}
			if err := validateSchema(schema, fixture); err != nil {
				t.Fatalf("fixture violates canonical schema: %v", err)
			}
			if err := json.Unmarshal([]byte(tt.fixture), tt.value); err != nil {
				t.Fatal(err)
			}
			switch value := tt.value.(type) {
			case *transport.HealthLiveResponse:
				if !value.Status.Valid() {
					t.Fatal("generated enum rejects canonical liveness state")
				}
			case *transport.HealthReadyResponse:
				if !value.Status.Valid() {
					t.Fatal("generated enum rejects canonical readiness state")
				}
				for _, check := range value.Checks {
					if !check.State.Valid() {
						t.Fatal("generated enum rejects canonical component state")
					}
				}
			}
			data, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatal(err)
			}
			var encoded any
			if err173 := json.Unmarshal(data, &encoded); err173 != nil {
				t.Fatal(err173)
			}
			if err176 := validateSchema(schema, encoded); err176 != nil {
				t.Fatalf("generated encoding violates canonical schema: %v", err176)
			}
			if !reflect.DeepEqual(encoded, fixture) {
				t.Fatalf("JSON round trip = %s, want %s", data, tt.fixture)
			}
		})
	}
}

func TestIdentityJSONMatchesCanonicalResponse(t *testing.T) {
	document := readJSON(t, "../../../../packages/openapi/openapi.json")
	schema := responseSchema(t, document, "/api/v1/me", "200", "IdentityResponse")
	fixture := []byte(`{"authenticated":true,"user":{"id":"00000000-0000-4000-8000-000000000001",` +
		`"email":"local@example.test"}}`)
	var value transport.TransportIdentityResponse
	if err := json.Unmarshal(fixture, &value); err != nil {
		t.Fatal(err)
	}
	if !value.Authenticated.Valid() {
		t.Fatal("generated enum rejects canonical identity")
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	var decoded any
	if decodeErr := json.Unmarshal(encoded, &decoded); decodeErr != nil {
		t.Fatal(decodeErr)
	}
	if schemaErr := validateSchema(schema, decoded); schemaErr != nil {
		t.Fatal(schemaErr)
	}
	for _, invalid := range []string{
		`{"authenticated":false,"user":{"id":"local","email":"local@example.test"}}`,
		`{"authenticated":true,"user":{"id":"local","email":42}}`,
		`{"authenticated":true,"user":{"id":"local","email":"local@example.test","subject":"provider-private"}}`,
		`{"authenticated":true}`,
	} {
		var candidate any
		if decodeErr := json.Unmarshal([]byte(invalid), &candidate); decodeErr != nil {
			t.Fatal(decodeErr)
		}
		if schemaErr := validateSchema(schema, candidate); schemaErr == nil {
			t.Fatal("canonical identity schema accepted invalid or private fields")
		}
	}
}

func TestInvalidStatesAndDiagnosticsFailCanonicalSchema(t *testing.T) {
	document := readJSON(t, "../../../../packages/openapi/openapi.json")
	for _, tt := range []struct{ path, status, name, fixture string }{
		{path: "/live", status: "200", name: "HealthLiveResponse", fixture: `{"status":"ready"}`},
		{path: "/live",
			status:  "200",
			name:    "HealthLiveResponse",
			fixture: `{"status":"alive","error":"private detail"}`},
		{path: "/ready",
			status:  "503",
			name:    "HealthReadyResponse",
			fixture: `{"status":"failed","checks":[]}`},
		{path: "/ready", status: "503", name: "HealthReadyResponse", fixture: `{"status":"not_ready"}`},
		{path: "/ready",
			status:  "503",
			name:    "HealthReadyResponse",
			fixture: `{"status":"not_ready","checks":null}`},
		{path: "/ready",
			status:  "503",
			name:    "HealthReadyResponse",
			fixture: `{"status":"not_ready","checks":[{"name":"database","state":"failed"}]}`},
		{path: "/ready",
			status:  "503",
			name:    "HealthReadyResponse",
			fixture: `{"status":"not_ready","checks":[{"name":"database","state":"not_ready","error":"private detail"}]}`},
	} {
		var value any
		if err := json.Unmarshal([]byte(tt.fixture), &value); err != nil {
			t.Fatal(err)
		}
		if err := validateSchema(responseSchema(t, document, tt.path, tt.status, tt.name), value); err == nil {
			t.Errorf("canonical schema accepted invalid fixture %s", tt.fixture)
		}
	}
	if (&transport.HealthLiveResponse{Status: "ready"}).Status.Valid() {
		t.Error("generated liveness enum accepts ready")
	}
	if (&transport.HealthReadyResponse{Status: "failed"}).Status.Valid() {
		t.Error("generated readiness enum accepts failed")
	}
}

//nolint:gocognit // Keep this regression scenario and its ordered failure assertions together.
func TestPinnedGenerationMatchesCheckedArtifact(t *testing.T) {
	lock := readJSON(t, "../../../../tools.lock.json")["oapi-codegen"].(map[string]any)
	module, version := lock["module"].(string), lock["version"].(string)
	if module != "github.com/oapi-codegen/oapi-codegen/v2" ||
		!strings.HasPrefix(version,
			"v2.") ||
		strings.ContainsAny(version,
			"@/ ") {
		t.Fatalf("invalid official generator pin: %s@%s", module, version)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	metadata, err := exec.CommandContext(ctx, "go", "mod", "download", "-json", module+"@"+version).Output()
	if err != nil {
		t.Fatalf("resolve pinned module: %v", err)
	}
	var download map[string]any
	if err246 := json.Unmarshal(metadata, &download); err246 != nil {
		t.Fatal(err246)
	}
	if download["Sum"] != lock["sum"] ||
		download["GoModSum"] != lock["goModSum"] ||
		download["Version"] != version {
		t.Fatalf("generator checksum/version differs from verified lock: %s", metadata)
	}
	checked, err := os.ReadFile("health.gen.go")
	if err != nil {
		t.Fatal(err)
	}
	marker := "Code generated by github.com/oapi-codegen/oapi-codegen/v2 version " + version + " DO NOT EDIT."
	if !bytes.Contains(checked, []byte(marker)) {
		t.Fatalf("generated artifact lacks pinned marker %q", marker)
	}
	for i := range 2 {
		output := filepath.Join(t.TempDir(), "health.gen.go")
		configData, err264 := os.ReadFile("oapi-codegen.yaml")
		if err264 != nil {
			t.Fatal(err264)
		}
		var config map[string]any
		if err269 := yaml.Unmarshal(configData, &config); err269 != nil {
			t.Fatal(err269)
		}
		// The generator gives the config's output precedence over the -o flag.
		// Change only the temporary output destination, preserving all options.
		config["output"] = output
		configData, err264 = yaml.Marshal(config)
		if err264 != nil {
			t.Fatal(err264)
		}
		configPath := filepath.Join(t.TempDir(), "oapi-codegen.yaml")
		if err280 := os.WriteFile(configPath, configData, 0o600); err280 != nil {
			t.Fatal(err280)
		}
		cmd := exec.CommandContext(ctx, "go", "run", module+"/cmd/oapi-codegen@"+version,
			"-config", configPath, "../../../../packages/openapi/openapi.json")
		if log, err285 := cmd.CombinedOutput(); err285 != nil {
			t.Fatalf("generation %d: %v\n%s", i+1, err285, log)
		}
		generated, err264 := os.ReadFile(output)
		if err264 != nil {
			t.Fatal(err264)
		}
		if !bytes.Equal(generated, checked) {
			t.Fatalf("generation %d differs from checked health.gen.go; regenerate using the locked module",
				i+1)
		}
	}
}
