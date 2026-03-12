// Package targets defines the registry of dependency directory types that
// Goclean can detect and clean.
package targets

// Target represents a type of dependency directory.
type Target struct {
	Name          string // Human-readable name (e.g. "Node.js")
	DependencyDir string // Directory name to look for (e.g. "node_modules")
	ConfigFile    string // Config file that lives alongside it (e.g. "package.json")
}

var registry = []Target{
	{Name: "Node.js", DependencyDir: "node_modules", ConfigFile: "package.json"},
	{Name: "PHP (Composer)", DependencyDir: "vendor", ConfigFile: "composer.json"},
	{Name: "Python (venv)", DependencyDir: ".venv", ConfigFile: "requirements.txt"},
	{Name: "Python (venv)", DependencyDir: "venv", ConfigFile: "requirements.txt"},
	{Name: "Rust", DependencyDir: "target", ConfigFile: "Cargo.toml"},
	{Name: "Java (Maven)", DependencyDir: "target", ConfigFile: "pom.xml"},
	{Name: "Go", DependencyDir: "vendor", ConfigFile: "go.mod"},
	{Name: "Dart/Flutter", DependencyDir: ".dart_tool", ConfigFile: "pubspec.yaml"},
	{Name: "CocoaPods", DependencyDir: "Pods", ConfigFile: "Podfile"},
	{Name: "Gradle", DependencyDir: "build", ConfigFile: "build.gradle"},
}

var byDir map[string]Target
var byDirAll map[string][]Target

func init() {
	byDir = make(map[string]Target, len(registry))
	byDirAll = make(map[string][]Target, len(registry))
	for _, t := range registry {
		if _, exists := byDir[t.DependencyDir]; !exists {
			byDir[t.DependencyDir] = t
		}
		byDirAll[t.DependencyDir] = append(byDirAll[t.DependencyDir], t)
	}
}

// All returns every registered target.
func All() []Target {
	out := make([]Target, len(registry))
	copy(out, registry)
	return out
}

// GetByDirName looks up the first target registered for a dependency directory name.
func GetByDirName(dir string) (Target, bool) {
	t, ok := byDir[dir]
	return t, ok
}

// GetAllByDirName returns all targets registered for a dependency directory name.
func GetAllByDirName(dir string) []Target {
	return byDirAll[dir]
}

// DirNames returns all unique dependency directory names.
func DirNames() []string {
	seen := make(map[string]bool, len(registry))
	var names []string
	for _, t := range registry {
		if !seen[t.DependencyDir] {
			seen[t.DependencyDir] = true
			names = append(names, t.DependencyDir)
		}
	}
	return names
}

// ForDirs returns the subset of targets whose DependencyDir is in the given set.
func ForDirs(enabled []string) []Target {
	set := make(map[string]bool, len(enabled))
	for _, d := range enabled {
		set[d] = true
	}
	var out []Target
	for _, t := range registry {
		if set[t.DependencyDir] {
			out = append(out, t)
		}
	}
	return out
}
