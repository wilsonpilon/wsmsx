package renum

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenumberGoldenPrograms(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		inputFile string
		wantFile  string
		opts      Options
		warnings  int
		wantErr   string
	}{
		{
			name:      "input_test34_on_strig_gosub_return",
			inputFile: filepath.Join("testdata", "golden", "test34.input.bas"),
			wantFile:  filepath.Join("testdata", "golden", "test34.expected.bas"),
			opts:      Options{StartLine: 100, Increment: 10, FromLine: 0},
		},
		{
			name:      "input_test72_restore_reset",
			inputFile: filepath.Join("testdata", "golden", "test72.input.bas"),
			wantFile:  filepath.Join("testdata", "golden", "test72.expected.bas"),
			opts:      Options{StartLine: 100, Increment: 10, FromLine: 0},
		},
		{
			name:      "input_test76_on_key_then_else_return",
			inputFile: filepath.Join("testdata", "golden", "test76.input.bas"),
			wantFile:  filepath.Join("testdata", "golden", "test76.expected.bas"),
			opts:      Options{StartLine: 100, Increment: 10, FromLine: 0},
		},
		{
			name:      "input_test33_function_keys",
			inputFile: filepath.Join("testdata", "golden", "test33.input.bas"),
			wantFile:  filepath.Join("testdata", "golden", "test33.expected.bas"),
			opts:      Options{StartLine: 100, Increment: 10, FromLine: 0},
		},
		{
			name:      "input_test71_restore_and_data",
			inputFile: filepath.Join("testdata", "golden", "test71.input.bas"),
			wantFile:  filepath.Join("testdata", "golden", "test71.expected.bas"),
			opts:      Options{StartLine: 100, Increment: 10, FromLine: 0},
		},
		{
			name:      "input_if_test1_branches",
			inputFile: filepath.Join("testdata", "golden", "if_test1.input.bas"),
			wantFile:  filepath.Join("testdata", "golden", "if_test1.expected.bas"),
			opts:      Options{StartLine: 100, Increment: 10, FromLine: 0},
		},
		{
			name:      "strict_parity_undefined_flow_error",
			inputFile: filepath.Join("testdata", "golden", "strict_undefined.input.bas"),
			opts:      Options{StartLine: 100, Increment: 10, FromLine: 0, StrictMSXParity: true},
			wantErr:   "Undefined line 999 in 10",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			in := mustReadGoldenFile(t, tc.inputFile)
			want := ""
			if tc.wantFile != "" {
				want = mustReadGoldenFile(t, tc.wantFile)
			}

			res, err := Renumber(in, tc.opts)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("Renumber() error = nil, want %q", tc.wantErr)
				}
				if err.Error() != tc.wantErr {
					t.Fatalf("Renumber() error mismatch\nwant:\n%q\ngot:\n%q", tc.wantErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("Renumber() error = %v", err)
			}
			if res.Text != want {
				t.Fatalf("Renumber() golden mismatch\nwant:\n%s\ngot:\n%s", want, res.Text)
			}
			if len(res.UndefinedRefs) != tc.warnings {
				t.Fatalf("UndefinedRefs length = %d, want %d", len(res.UndefinedRefs), tc.warnings)
			}
		})
	}
}

func mustReadGoldenFile(t *testing.T, relativePath string) string {
	t.Helper()
	data, err := os.ReadFile(relativePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", relativePath, err)
	}
	return string(data)
}
