package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func initRepo(t *testing.T, dir string) {
	t.Helper()
	run(t, dir, "git", "init", "-b", "main")
	run(t, dir, "git", "config", "user.email", "test@test.com")
	run(t, dir, "git", "config", "user.name", "Test")
}

func run(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v failed in %s: %v\n%s", name, args, dir, err, out)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
}

func TestGetChangedFiles_WithExplicitFiles(t *testing.T) {
	cd := NewChangeDetectorWithOptions(ChangeOptions{
		Files: []string{"a.go", "b.go", "a.go", " c.go "},
	})

	files, err := cd.GetChangedFiles()
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d: %v", len(files), files)
	}
	if files[0] != "a.go" || files[1] != "b.go" || files[2] != "c.go" {
		t.Errorf("unexpected files: %v", files)
	}
}

func TestGetChangedFiles_Uncommitted(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	chdir(t, dir)

	writeTestFile(t, filepath.Join(dir, "file.txt"), "initial")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "init")

	writeTestFile(t, filepath.Join(dir, "file.txt"), "modified")

	cd := NewChangeDetectorWithOptions(ChangeOptions{Uncommitted: true})
	files, err := cd.GetChangedFiles()
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 || files[0] != "file.txt" {
		t.Errorf("expected [file.txt], got %v", files)
	}
}

func TestGetChangedFiles_UncommittedStaged(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	chdir(t, dir)

	writeTestFile(t, filepath.Join(dir, "file.txt"), "initial")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "init")

	writeTestFile(t, filepath.Join(dir, "staged.txt"), "new")
	run(t, dir, "git", "add", "staged.txt")

	cd := NewChangeDetectorWithOptions(ChangeOptions{Uncommitted: true})
	files, err := cd.GetChangedFiles()
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 || files[0] != "staged.txt" {
		t.Errorf("expected [staged.txt], got %v", files)
	}
}

func TestGetChangedFiles_Untracked(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	chdir(t, dir)

	writeTestFile(t, filepath.Join(dir, "tracked.txt"), "initial")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "init")

	writeTestFile(t, filepath.Join(dir, "newfile.txt"), "untracked")

	cd := NewChangeDetectorWithOptions(ChangeOptions{Untracked: true})
	files, err := cd.GetChangedFiles()
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 || files[0] != "newfile.txt" {
		t.Errorf("expected [newfile.txt], got %v", files)
	}
}

func TestGetChangedFiles_BaseAndHead(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	chdir(t, dir)

	writeTestFile(t, filepath.Join(dir, "base.txt"), "base")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "base commit")

	run(t, dir, "git", "checkout", "-b", "feature")
	writeTestFile(t, filepath.Join(dir, "feature.txt"), "feature")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "feature commit")

	cd := NewChangeDetectorWithOptions(ChangeOptions{
		Base: "main",
		Head: "feature",
	})
	files, err := cd.GetChangedFiles()
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 || files[0] != "feature.txt" {
		t.Errorf("expected [feature.txt], got %v", files)
	}
}

func TestGetChangedFiles_BaseOnly(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	chdir(t, dir)

	writeTestFile(t, filepath.Join(dir, "base.txt"), "base")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "base commit")

	run(t, dir, "git", "checkout", "-b", "work")
	writeTestFile(t, filepath.Join(dir, "committed.txt"), "committed")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "work commit")

	writeTestFile(t, filepath.Join(dir, "uncommitted.txt"), "uncommitted")
	writeTestFile(t, filepath.Join(dir, "untracked.txt"), "untracked")
	run(t, dir, "git", "add", "uncommitted.txt")

	cd := NewChangeDetectorWithOptions(ChangeOptions{Base: "main"})
	files, err := cd.GetChangedFiles()
	if err != nil {
		t.Fatal(err)
	}

	fileSet := make(map[string]bool)
	for _, f := range files {
		fileSet[f] = true
	}

	if !fileSet["committed.txt"] {
		t.Error("expected committed.txt in changed files")
	}
	if !fileSet["uncommitted.txt"] {
		t.Error("expected uncommitted.txt in changed files")
	}
	if !fileSet["untracked.txt"] {
		t.Error("expected untracked.txt in changed files")
	}
}

func TestGetChangedFiles_DefaultBase(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	chdir(t, dir)

	run(t, dir, "git", "checkout", "-b", "main")
	writeTestFile(t, filepath.Join(dir, "base.txt"), "base")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "init on main")

	run(t, dir, "git", "checkout", "-b", "feature")
	writeTestFile(t, filepath.Join(dir, "new.txt"), "new")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "feature")

	cd := NewChangeDetectorWithOptions(ChangeOptions{})
	files, err := cd.GetChangedFiles()
	if err != nil {
		t.Fatal(err)
	}

	fileSet := make(map[string]bool)
	for _, f := range files {
		fileSet[f] = true
	}

	if !fileSet["new.txt"] {
		t.Error("expected new.txt in changed files (default base=main)")
	}
}

func TestNormalizeFiles_Deduplication(t *testing.T) {
	input := []string{"b.go", "a.go", "b.go", " c.go ", "", "a.go"}
	result := normalizeFiles(input)

	if len(result) != 3 {
		t.Fatalf("expected 3 files, got %d: %v", len(result), result)
	}
	if result[0] != "a.go" || result[1] != "b.go" || result[2] != "c.go" {
		t.Errorf("expected [a.go b.go c.go], got %v", result)
	}
}

func TestIsPathChanged_PrefixMatch(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	chdir(t, dir)

	writeTestFile(t, filepath.Join(dir, "infra", "infra-1", "main.tf"), "resource {}")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "init")

	run(t, dir, "git", "checkout", "-b", "change")
	writeTestFile(t, filepath.Join(dir, "infra", "infra-1", "main.tf"), "resource updated {}")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "change")

	cd := NewChangeDetectorWithOptions(ChangeOptions{Base: "main", Head: "change"})

	changed, err := cd.IsPathChanged("infra/infra-1")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Error("expected infra/infra-1 to be changed")
	}

	changed, err = cd.IsPathChanged("apps")
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("expected apps to not be changed")
	}
}

func TestIsPathChanged_RootPath(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	chdir(t, dir)

	writeTestFile(t, filepath.Join(dir, "file.txt"), "initial")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "init")

	writeTestFile(t, filepath.Join(dir, "file.txt"), "modified")

	cd := NewChangeDetectorWithOptions(ChangeOptions{Uncommitted: true})

	changed, err := cd.IsPathChanged("")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Error("expected root path to report changes")
	}

	changed, err = cd.IsPathChanged("./")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Error("expected ./ path to report changes")
	}
}

func TestIsIntentFileChanged(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	chdir(t, dir)

	writeTestFile(t, filepath.Join(dir, "examples", "intent.yaml"), "kind: Intent")
	writeTestFile(t, filepath.Join(dir, "other.txt"), "other")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "init")

	run(t, dir, "git", "checkout", "-b", "change")
	writeTestFile(t, filepath.Join(dir, "examples", "intent.yaml"), "kind: Intent\nmodified: true")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "change intent")

	cd := NewChangeDetectorWithOptions(ChangeOptions{Base: "main", Head: "change"})

	changed, err := cd.IsIntentFileChanged("intent.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Error("expected intent.yaml (basename) to match examples/intent.yaml")
	}

	changed, err = cd.IsIntentFileChanged("examples/intent.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Error("expected examples/intent.yaml to match")
	}

	changed, err = cd.IsIntentFileChanged("other-intent.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("expected other-intent.yaml to not match")
	}
}

func TestValidateOptions_Conflicts(t *testing.T) {
	tests := []struct {
		name    string
		options ChangeOptions
		wantErr bool
	}{
		{
			name:    "files with uncommitted",
			options: ChangeOptions{Files: []string{"a.go"}, Uncommitted: true},
			wantErr: true,
		},
		{
			name:    "files with base",
			options: ChangeOptions{Files: []string{"a.go"}, Base: "main"},
			wantErr: true,
		},
		{
			name:    "uncommitted with untracked",
			options: ChangeOptions{Uncommitted: true, Untracked: true},
			wantErr: true,
		},
		{
			name:    "uncommitted with base",
			options: ChangeOptions{Uncommitted: true, Base: "main"},
			wantErr: true,
		},
		{
			name:    "untracked with base",
			options: ChangeOptions{Untracked: true, Base: "main"},
			wantErr: true,
		},
		{
			name:    "head without base",
			options: ChangeOptions{Head: "feature"},
			wantErr: true,
		},
		{
			name:    "valid files only",
			options: ChangeOptions{Files: []string{"a.go"}},
			wantErr: false,
		},
		{
			name:    "valid base and head",
			options: ChangeOptions{Base: "main", Head: "feature"},
			wantErr: false,
		},
		{
			name:    "valid base only",
			options: ChangeOptions{Base: "main"},
			wantErr: false,
		},
		{
			name:    "valid uncommitted only",
			options: ChangeOptions{Uncommitted: true},
			wantErr: false,
		},
		{
			name:    "valid untracked only",
			options: ChangeOptions{Untracked: true},
			wantErr: false,
		},
		{
			name:    "empty options",
			options: ChangeOptions{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateOptions(tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateOptions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetChangedFilesUnderPath(t *testing.T) {
	dir := t.TempDir()
	initRepo(t, dir)
	chdir(t, dir)

	writeTestFile(t, filepath.Join(dir, "infra", "a.tf"), "a")
	writeTestFile(t, filepath.Join(dir, "apps", "b.go"), "b")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "init")

	run(t, dir, "git", "checkout", "-b", "change")
	writeTestFile(t, filepath.Join(dir, "infra", "a.tf"), "a modified")
	writeTestFile(t, filepath.Join(dir, "apps", "b.go"), "b modified")
	run(t, dir, "git", "add", ".")
	run(t, dir, "git", "commit", "-m", "change")

	cd := NewChangeDetectorWithOptions(ChangeOptions{Base: "main", Head: "change"})

	files, err := cd.GetChangedFilesUnderPath("infra")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != "infra/a.tf" {
		t.Errorf("expected [infra/a.tf], got %v", files)
	}

	files, err = cd.GetChangedFilesUnderPath("charts")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Errorf("expected empty, got %v", files)
	}
}
