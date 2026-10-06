package transport_test

import (
	"bytes"
	"context"
	"encoding/json"
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
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	return result
}

// Check the schema from the canonical document, rather than a second health schema.
// Fail closed if this small fixture validator encounters an unsupported schema type.
func validateSchema(schema map[string]any, value any) error {
	switch schema["type"] {
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

func TestHealthJSONMatchesCanonicalResponses(t *testing.T) {
	document := readJSON(t, "../../../../packages/openapi/openapi.json")
	tests := []struct {
		name, path, status, schema, fixture string
		value                               any
	}{
		{"live", "/live", "200", "HealthLiveResponse", `{"status":"alive"}`, &transport.HealthLiveResponse{Status: "alive"}},
		{"ready", "/ready", "200", "HealthReadyResponse", `{"status":"ready","checks":[{"name":"database","state":"ready"}]}`, &transport.HealthReadyResponse{}},
		{"not_ready", "/ready", "503", "HealthReadyResponse", `{"status":"not_ready","checks":[{"name":"database","state":"not_ready"},{"name":"redis","state":"ready"}]}`, &transport.HealthReadyResponse{}},
		{"no_dependencies", "/ready", "200", "HealthReadyResponse", `{"status":"ready","checks":[]}`, &transport.HealthReadyResponse{}},
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
			if err := json.Unmarshal(data, &encoded); err != nil {
				t.Fatal(err)
			}
			if err := validateSchema(schema, encoded); err != nil {
				t.Fatalf("generated encoding violates canonical schema: %v", err)
			}
			if !reflect.DeepEqual(encoded, fixture) {
				t.Fatalf("JSON round trip = %s, want %s", data, tt.fixture)
			}
		})
	}
}

func TestInvalidStatesAndDiagnosticsFailCanonicalSchema(t *testing.T) {
	document := readJSON(t, "../../../../packages/openapi/openapi.json")
	for _, tt := range []struct{ path, status, name, fixture string }{
		{"/live", "200", "HealthLiveResponse", `{"status":"ready"}`},
		{"/live", "200", "HealthLiveResponse", `{"status":"alive","error":"private detail"}`},
		{"/ready", "503", "HealthReadyResponse", `{"status":"failed","checks":[]}`},
		{"/ready", "503", "HealthReadyResponse", `{"status":"not_ready"}`},
		{"/ready", "503", "HealthReadyResponse", `{"status":"not_ready","checks":null}`},
		{"/ready", "503", "HealthReadyResponse", `{"status":"not_ready","checks":[{"name":"database","state":"failed"}]}`},
		{"/ready", "503", "HealthReadyResponse", `{"status":"not_ready","checks":[{"name":"database","state":"not_ready","error":"private detail"}]}`},
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

func TestPinnedGenerationMatchesCheckedArtifact(t *testing.T) {
	lock := readJSON(t, "../../../../tools.lock.json")["oapi-codegen"].(map[string]any)
	module, version := lock["module"].(string), lock["version"].(string)
	if module != "github.com/oapi-codegen/oapi-codegen/v2" || !strings.HasPrefix(version, "v2.") || strings.ContainsAny(version, "@/ ") {
		t.Fatalf("invalid official generator pin: %s@%s", module, version)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	metadata, err := exec.CommandContext(ctx, "go", "mod", "download", "-json", module+"@"+version).Output()
	if err != nil {
		t.Fatalf("resolve pinned module: %v", err)
	}
	var download map[string]any
	if err := json.Unmarshal(metadata, &download); err != nil {
		t.Fatal(err)
	}
	if download["Sum"] != lock["sum"] || download["GoModSum"] != lock["goModSum"] || download["Version"] != version {
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
		configData, err := os.ReadFile("oapi-codegen.yaml")
		if err != nil {
			t.Fatal(err)
		}
		var config map[string]any
		if err := yaml.Unmarshal(configData, &config); err != nil {
			t.Fatal(err)
		}
		// The generator gives the config's output precedence over the -o flag.
		// Change only the temporary output destination, preserving all options.
		config["output"] = output
		configData, err = yaml.Marshal(config)
		if err != nil {
			t.Fatal(err)
		}
		configPath := filepath.Join(t.TempDir(), "oapi-codegen.yaml")
		if err := os.WriteFile(configPath, configData, 0o600); err != nil {
			t.Fatal(err)
		}
		cmd := exec.CommandContext(ctx, "go", "run", module+"/cmd/oapi-codegen@"+version,
			"-config", configPath, "../../../../packages/openapi/openapi.json")
		if log, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("generation %d: %v\n%s", i+1, err, log)
		}
		generated, err := os.ReadFile(output)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(generated, checked) {
			t.Fatalf("generation %d differs from checked health.gen.go; regenerate using the locked module", i+1)
		}
	}
}
