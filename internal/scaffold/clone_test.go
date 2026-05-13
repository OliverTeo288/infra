package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsDirEmpty(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		want    bool
		wantErr bool
	}{
		{
			name:  "truly empty",
			setup: func(t *testing.T) string { return t.TempDir() },
			want:  true,
		},
		{
			name: "only .git is empty",
			setup: func(t *testing.T) string {
				d := t.TempDir()
				if err := os.Mkdir(filepath.Join(d, ".git"), 0o755); err != nil {
					t.Fatal(err)
				}
				return d
			},
			want: true,
		},
		{
			name: "has a regular file",
			setup: func(t *testing.T) string {
				d := t.TempDir()
				if err := os.WriteFile(filepath.Join(d, "README.md"), []byte("x"), 0o644); err != nil {
					t.Fatal(err)
				}
				return d
			},
			want: false,
		},
		{
			name: ".DS_Store only is empty",
			setup: func(t *testing.T) string {
				d := t.TempDir()
				if err := os.WriteFile(filepath.Join(d, ".DS_Store"), []byte("x"), 0o644); err != nil {
					t.Fatal(err)
				}
				return d
			},
			want: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := isDirEmpty(tc.setup(t))
			if (err != nil) != tc.wantErr {
				t.Fatalf("isDirEmpty err = %v; wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("isDirEmpty = %v; want %v", got, tc.want)
			}
		})
	}
}

func TestRepoNameFromURL(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"https://gitlab.example.com/group/subgroup/project.git", "project"},
		{"https://gitlab.example.com/group/subgroup/project", "project"},
		{"git@gitlab.example.com:group/subgroup/project.git", "project"},
		{"https://gitlab.example.com/group/sub/project.git?ref=main", "project"},
		{"https://gitlab.example.com/group/sub/project.git#readme", "project"},
		{"", "template-repo"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			if got := repoNameFromURL(tc.in); got != tc.want {
				t.Errorf("repoNameFromURL(%q) = %q; want %q", tc.in, got, tc.want)
			}
		})
	}
}
