// Package targets definisce il registro dei tipi di cartelle di dipendenze
// che Corgab è in grado di rilevare e pulire.
package targets

// Target rappresenta un tipo di cartella di dipendenze.
// Ogni target associa un ecosistema (es. Node.js) alla cartella generata
// dal package manager (es. node_modules) e al file di configurazione
// che ne attesta la presenza (es. package.json).
type Target struct {
	Name          string // Nome leggibile dell'ecosistema (es. "Node.js")
	DependencyDir string // Nome della cartella di dipendenze da cercare (es. "node_modules")
	ConfigFile    string // File di configurazione presente nella directory del progetto (es. "package.json")
}

// registry contiene tutti i target supportati.
// Nota: alcune cartelle (vendor, target) sono condivise da più ecosistemi.
// La disambiguazione avviene nel package scanner verificando quale config file esiste.
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

// byDir mappa nome cartella -> primo target registrato (lookup veloce).
var byDir map[string]Target

// byDirAll mappa nome cartella -> tutti i target registrati (per disambiguazione).
var byDirAll map[string][]Target

// init popola le mappe di lookup all'avvio del programma.
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

// All restituisce una copia di tutti i target registrati.
func All() []Target {
	out := make([]Target, len(registry))
	copy(out, registry)
	return out
}

// GetByDirName restituisce il primo target associato al nome di cartella specificato.
func GetByDirName(dir string) (Target, bool) {
	t, ok := byDir[dir]
	return t, ok
}

// GetAllByDirName restituisce tutti i target associati al nome di cartella specificato.
// Usato per disambiguare cartelle condivise (es. "vendor" -> PHP e Go).
func GetAllByDirName(dir string) []Target {
	return byDirAll[dir]
}

// DirNames restituisce tutti i nomi univoci delle cartelle di dipendenze.
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

// ForDirs restituisce il sottoinsieme di target le cui cartelle sono nell'elenco fornito.
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
