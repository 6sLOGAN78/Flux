package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func forbiddenVendor(path string) bool {
	if strings.HasPrefix(path, "github.com/newrelic/") {
		return true
	}
	for _, name := range []string{"nrecho", "nrpgx", "nrredis", "nrpkgerrors"} {
		if strings.Contains(path, name) {
			return true
		}
	}
	return false
}

func inspectVendorGraph(data []byte) error {
	d := json.NewDecoder(bytes.NewReader(data))
	count := 0
	for {
		var entry struct {
			ImportPath string
			Path       string
			Imports    []string
			Module     *struct {
				Path    string
				Replace *struct{ Path string }
			}
			Replace    *struct{ Path string }
			Error      json.RawMessage
			DepsErrors []json.RawMessage
		}
		if err := d.Decode(&entry); err != nil {
			if err == io.EOF && count > 0 {
				return nil
			}
			return fmt.Errorf("invalid dependency output: %w", err)
		}
		count++
		if len(entry.Error) > 0 || len(entry.DepsErrors) > 0 {
			return fmt.Errorf("dependency resolution failed")
		}
		paths := append(entry.Imports, entry.ImportPath, entry.Path)
		if entry.Module != nil {
			paths = append(paths, entry.Module.Path)
			if entry.Module.Replace != nil {
				paths = append(paths, entry.Module.Replace.Path)
			}
		}
		if entry.Replace != nil {
			paths = append(paths, entry.Replace.Path)
		}
		for _, path := range paths {
			if forbiddenVendor(path) {
				return fmt.Errorf("forbidden dependency: %s", path)
			}
		}
	}
}

func inspectVendorSource(root string) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == ".git" || entry.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range file.Imports {
			name, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				return err
			}
			if forbiddenVendor(name) {
				return fmt.Errorf("forbidden authored import: %s: %s", path, name)
			}
		}
		return nil
	})
}

func TestForbiddenVendorDependencies(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	for _, args := range [][]string{
		{"list", "-deps", "-json", "./cmd/api", "./cmd/redirector", "./cmd/worker", "./cmd/migrator"},
		{"list", "-m", "-json", "all"},
	} {
		cmd := exec.CommandContext(ctx, "go", args...)
		cmd.Dir = root
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		data, err := cmd.Output()
		if err != nil {
			t.Fatalf("dependency tool failed: %v: %s", err, stderr.String())
		}
		if err := inspectVendorGraph(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := inspectVendorSource(root); err != nil {
		t.Fatal(err)
	}
}

func TestForbiddenVendorGraphRejectsHiddenDependencies(t *testing.T) {
	for _, data := range []string{
		`{"ImportPath":"github.com/newrelic/go-agent/v3/newrelic"}`,
		`{"Path":"github.com/newrelic/go-agent/v3"}`,
		`{"ImportPath":"safe","Module":{"Path":"github.com/newrelic/module"}}`,
		`{"Path":"safe","Replace":{"Path":"github.com/newrelic/module"}}`,
		`{"Imports":["example.com/nrredis"]}`, `{"Error":{"Err":"failed"}}`, `garbage`, "",
	} {
		if inspectVendorGraph([]byte(data)) == nil {
			t.Fatalf("accepted %q", data)
		}
	}
	if err := inspectVendorGraph([]byte(`{"Path":"go.opentelemetry.io/otel"}`)); err != nil {
		t.Fatal(err)
	}
}
