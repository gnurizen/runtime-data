package integration

import (
	"bytes"
	"debug/elf"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v3"

	"github.com/parca-dev/runtime-data/pkg/datamap"
	"github.com/parca-dev/runtime-data/pkg/python"
)

var update = flag.Bool("update", false, "update golden files")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

var pythonVersions = []string{
	"2.7.15",
	"3.3.7",
	"3.4.8",
	"3.5.5",
	"3.6.6",
	"3.7.0",
	"3.8.0",
	"3.9.5",
	"3.10.0",
	"3.11.0",
}

func TestPythonIntegration(t *testing.T) {
	t.Parallel()

	for _, version := range pythonVersions {
		version := version
		t.Run(version, func(t *testing.T) {
			t.Parallel()

			layoutMap := python.DataMapForVersion(version)
			if layoutMap == nil {
				t.Fatalf("python.DataMapForVersion(%s) = nil", version)
			}

			dm, err := datamap.New(layoutMap)
			if err != nil {
				t.Fatalf("python.GenerateDataMap(%s) = %v", version, err)
			}

			parts := strings.Split(version, ".")
			matches, err := filepath.Glob(fmt.Sprintf("tmp/libpython%s.%s*.so.1.0", parts[0], parts[1]))
			if err != nil {
				t.Fatalf("filepath.Glob() = %v", err)
			}
			if len(matches) == 0 {
				t.Fatalf("filepath.Glob() = no matches")
			}
			input := matches[0]

			f, err := elf.Open(input)
			if err != nil {
				t.Fatalf("elf.Open() = %v", err)
			}

			dwarfData, err := f.DWARF()
			if err != nil {
				t.Fatalf("f.DWARF() = %v", err)
			}

			if err := dm.ReadFromDWARF(dwarfData); err != nil {
				t.Errorf("input: %s", input)
				t.Fatalf("datamap.ReadFromDWARF() = %v", err)
			}

			got := layoutMap.Layout().(*python.Layout)

			golden := filepath.Join("testdata", fmt.Sprintf("python_%s.yaml", sanitizeIdentifier(version)))
			if *update {
				var buf bytes.Buffer
				enc := yaml.NewEncoder(&buf)
				enc.SetIndent(2)
				if err := enc.Encode(got); err != nil {
					t.Fatalf("yaml.Encode() = %v", err)
				}
				if err := os.WriteFile(golden, buf.Bytes(), 0o644); err != nil {
					t.Fatalf("os.WriteFile() = %v", err)
				}
			}

			wantData, err := os.ReadFile(golden)
			if err != nil {
				t.Fatalf("os.ReadFile() = %v", err)
			}

			var want python.Layout
			yaml.Unmarshal(wantData, &want)

			if diff := cmp.Diff(want, *got, cmp.AllowUnexported(python.Layout{})); diff != "" {
				t.Errorf("input: %s, golden: %s", input, golden)
				t.Errorf("python.GenerateDataMap(%s) mismatch (-want +got):\n%s", version, diff)
			}
		})
	}
}

// sanitizeIdentifier sanitizes the identifier to be used as a filename.
func sanitizeIdentifier(identifier string) string {
	return strings.ReplaceAll(identifier, ".", "_")
}
