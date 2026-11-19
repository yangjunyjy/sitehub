package architecture_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

func TestModulesOnlyCrossPublicContracts(t *testing.T) {
	_, source, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(source), "..", "..", "service")
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		relative, _ := filepath.Rel(root, path)
		module := strings.Split(filepath.ToSlash(relative), "/")[0]
		partsOfPath := strings.Split(filepath.ToSlash(relative), "/")
		isAPI := len(partsOfPath) > 1 && partsOfPath[1] == "api"
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			dep, _ := strconv.Unquote(spec.Path.Value)
			if strings.HasPrefix(dep, "sitehub/service/") {
				parts := strings.Split(strings.TrimPrefix(dep, "sitehub/service/"), "/")
				if parts[0] == "file" || parts[0] == "scheduler" {
					t.Errorf("%s depends on a service merged into blog: %s", relative, dep)
				}
				publicRPC := len(parts) > 2 && parts[1] == "rpc" && (parts[2] == "pb" || strings.HasSuffix(parts[2], "_client"))
				if parts[0] != module && (len(parts) < 2 || parts[1] != "contract" && parts[1] != "client") && !publicRPC {
					t.Errorf("%s imports another module's implementation: %s", relative, dep)
				}
				if isAPI && module == "rbac" && !publicRPC && (len(parts) < 2 || (parts[1] != "api" && parts[1] != "contract")) {
					t.Errorf("%s RBAC API must call its RPC boundary: %s", relative, dep)
				}
			}
			if isAPI && !strings.Contains(filepath.ToSlash(relative), "/api/internal/svc/") && (dep == "database/sql" || strings.HasPrefix(dep, "gorm.io/") || strings.Contains(dep, "/stores/sqlx") || strings.Contains(dep, "/infrastructure/database")) {
				t.Errorf("%s API opens a persistence boundary: %s", relative, dep)
			}
			if dep == "github.com/gin-gonic/gin" || (!isAPI && strings.Contains(dep, "common/http/")) {
				t.Errorf("%s business/RPC layer depends on an HTTP adapter: %s", relative, dep)
			}
			if dep == "sitehub/internal/services" || strings.HasPrefix(dep, "sitehub/internal/repository") {
				t.Errorf("%s reintroduces legacy shared business packages", relative)
			}
			if module == "blog" && len(partsOfPath) > 1 && partsOfPath[1] == "jobs" && (strings.Contains(dep, "/model") || strings.Contains(dep, "/repository") || strings.Contains(dep, "gorm.io/") || strings.Contains(dep, "redis") || dep == "database/sql") {
				t.Errorf("scheduler must use capabilities, not data access: %s", dep)
			}
			if strings.HasPrefix(dep, "sitehub/compat/") {
				t.Errorf("%s depends on HTTP/application assembly: %s", relative, dep)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMergedServicesHaveNoSeparateDirectories(t *testing.T) {
	_, source, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(source), "..", "..", "service")
	for _, retired := range []string{"file", "scheduler", "blog/rpc", "product/rpc"} {
		if _, err := os.Stat(filepath.Join(root, retired)); !os.IsNotExist(err) {
			t.Errorf("%s must live inside blog, without a separate service directory (stat: %v)", retired, err)
		}
	}
}

func TestBlogHasOneAssemblyForEachLayer(t *testing.T) {
	_, source, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(source), "..", "..", "service", "blog")
	declarations := map[string][]string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		relative, _ := filepath.Rel(root, path)
		relative = filepath.ToSlash(relative)
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		for _, declaration := range file.Decls {
			group, ok := declaration.(*ast.GenDecl)
			if !ok || group.Tok != token.TYPE {
				continue
			}
			for _, spec := range group.Specs {
				typeName := spec.(*ast.TypeSpec).Name.Name
				if typeName == "Module" {
					declarations["Module"] = append(declarations["Module"], relative)
				}
				layer := strings.Split(relative, "/")[0]
				if (layer == "api" || layer == "rpc") && (typeName == "Config" || typeName == "ServiceContext") {
					key := layer + "/" + typeName
					declarations[key] = append(declarations[key], relative)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"Module", "api/Config", "api/ServiceContext"} {
		if paths := declarations[key]; len(paths) != 1 {
			t.Errorf("blog must have one %s, found %d: %v", key, len(paths), paths)
		}
	}
}

func TestCommonDoesNotImportBusinessImplementations(t *testing.T) {
	_, source, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(source), "..", "..", "common")
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range parsed.Imports {
			dep, _ := strconv.Unquote(spec.Path.Value)
			if dep == "github.com/gin-gonic/gin" || strings.HasPrefix(dep, "sitehub/service/") || strings.HasPrefix(dep, "sitehub/compat/") {
				t.Errorf("common must not depend on business or compatibility packages: %s imports %s", path, dep)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestControlPlaneConsumersDoNotImportProviderSDKs(t *testing.T) {
	_, source, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(source), "..", "..")
	for _, relative := range []string{"common/registry", "common/configcenter", "common/governance", "common/command", "gateway/internal/discovery"} {
		err := filepath.WalkDir(filepath.Join(root, relative), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, spec := range parsed.Imports {
				dep, _ := strconv.Unquote(spec.Path.Value)
				if strings.HasPrefix(dep, "go.etcd.io/") || strings.HasPrefix(dep, "sitehub/common/controlplane/etcd") {
					t.Errorf("%s must use the registry/config-center contracts instead of a provider implementation: %s", relative, dep)
				}
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestNoCompatibilityRuntime(t *testing.T) {
	_, source, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(source), "..", "..")
	if _, err := os.Stat(filepath.Join(root, "compat")); !os.IsNotExist(err) {
		t.Fatal("the new project must not contain a compatibility runtime")
	}
	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"github.com/gin-gonic/gin", "my-blog-backend"} {
		if strings.Contains(string(mod), forbidden) {
			t.Fatalf("forbidden module dependency: %s", forbidden)
		}
	}
}
