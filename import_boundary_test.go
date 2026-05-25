package jules

import (
	"os/exec"
	"strings"
	"testing"
)

func TestPackageHasNoAppImports(t *testing.T) {
	output, err := exec.Command("go", "list", "-f", "{{join .Imports \"\\n\"}}", ".").Output()
	if err != nil {
		t.Fatalf("go list imports: %v", err)
	}

	forbidden := []string{
		"/" + "internal" + "/",
		"github.com/" + "sp" + "f13/cobra",
		"github.com/" + "sp" + "f13/viper",
		"github.com/" + "google/go" + "-github",
		"github.com/" + "model" + "contextprotocol/",
		"google.golang.org/" + "ge" + "nai",
		"os/exec",
		"path/filepath",
	}

	imports := strings.SplitSeq(strings.TrimSpace(string(output)), "\n")
	for importPath := range imports {
		for _, forbiddenImport := range forbidden {
			if strings.Contains(importPath, forbiddenImport) {
				t.Fatalf("go-jules imports forbidden dependency %q via %q", forbiddenImport, importPath)
			}
		}
	}
}
