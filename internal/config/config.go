// Package config gestisce il caricamento e il salvataggio del file di
// configurazione YAML di Corgab (~/.corgab.yaml).
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Config contiene le preferenze dell'utente per Corgab.
type Config struct {
	Days          int      `yaml:"days"`           // Soglia di giorni di inattività
	Targets       []string `yaml:"targets"`        // Nomi delle cartelle di dipendenze da monitorare
	ExcludedPaths []string `yaml:"excluded_paths"` // Percorsi da ignorare durante la scansione
}

// Default restituisce una configurazione con valori predefiniti e nessun target selezionato.
// I target vengono scelti dall'utente al primo avvio tramite il wizard.
func Default() Config {
	return Config{
		Days:          30,
		Targets:       []string{},
		ExcludedPaths: []string{},
	}
}

// Exists verifica se esiste un file di configurazione al percorso indicato.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Load legge e deserializza il file YAML di configurazione.
func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Save serializza la configurazione in formato YAML e la scrive su disco.
func Save(cfg Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
