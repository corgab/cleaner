# corgab cleaner

Un tool CLI interattivo veloce che scansiona il tuo filesystem alla ricerca di cartelle di dipendenze inutilizzate (`node_modules`, `vendor`, `.venv`, ecc.) e ti permette di eliminarle in blocco per recuperare spazio su disco.

Scritto in Go con scansione parallela e interfaccia terminale [Bubbletea](https://github.com/charmbracelet/bubbletea).

## Funzionalità

- **Rilevamento intelligente** — segnala solo i progetti il cui file di configurazione (`package.json`, `composer.json`, ecc.) non è stato modificato da N giorni
- **Scansione parallela** — usa un pool limitato di goroutine per scansionare filesystem anche molto grandi
- **TUI interattiva** — naviga i risultati, seleziona/deseleziona con la barra spaziatrice, vedi lo spazio recuperabile in tempo reale
- **Wizard di primo avvio** — scegli quali tipi di dipendenze monitorare al primo utilizzo
- **Modalità dry-run** — simula l'eliminazione senza rimuovere nulla
- **Statistiche cumulative** — tiene traccia di quanto spazio hai liberato nel tempo
- **Configurabile** — soglia giorni, percorsi esclusi e selezione target salvati in `~/.corgab.yaml`
- **Sicuro** — salta automaticamente le directory nascoste (`.git`, ecc.) e i percorsi di sistema (`/System`, `/Library`, ecc.)

## Target Supportati

| Ecosistema    | Cartella Dipendenze | File di Configurazione |
|--------------|--------------------|-----------------------|
| Node.js      | `node_modules`     | `package.json`        |
| PHP          | `vendor`           | `composer.json`       |
| Python       | `.venv` / `venv`   | `requirements.txt`    |
| Rust         | `target`           | `Cargo.toml`          |
| Java Maven   | `target`           | `pom.xml`             |
| Go           | `vendor`           | `go.mod`              |
| Dart/Flutter | `.dart_tool`       | `pubspec.yaml`        |
| CocoaPods    | `Pods`             | `Podfile`             |
| Gradle       | `build`            | `build.gradle`        |

I target sono opt-in: scegli quali monitorare durante il wizard di configurazione.

## Requisiti

- **Go 1.22+** — installa con `brew install go` (macOS) oppure visita [go.dev/dl](https://go.dev/dl)
- Dopo l'installazione di Go, aggiungi `~/go/bin` al tuo PATH se non è già presente:
  ```bash
  echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.zshrc
  source ~/.zshrc
  ```

## Installazione

### Da sorgente

```bash
git clone https://github.com/corgab/cleaner.git
cd cleaner
go install ./cmd/corgab
```

Ora il comando `corgab` è disponibile globalmente da qualsiasi directory.

### Build locale (senza installazione globale)

```bash
go build -o corgab ./cmd/corgab
./corgab
```

## Utilizzo

### Base

```bash
# Scansiona la home directory con le impostazioni predefinite (soglia 30 giorni)
corgab
```

Al primo avvio, il wizard ti chiederà quali tipi di dipendenze monitorare.

### Flag

| Flag             | Default   | Descrizione                                        |
|------------------|-----------|----------------------------------------------------|
| `--path <dir>`   | `~`       | Directory radice da scansionare                    |
| `--days <n>`     | `30`      | Mostra progetti inattivi da più di N giorni        |
| `--dry-run`      | `false`   | Simula l'eliminazione senza rimuovere nulla        |
| `--stats`        | `false`   | Mostra le statistiche cumulative ed esci           |
| `--reset-config` | `false`   | Riesegui il wizard di configurazione               |

### Esempi

```bash
# Scansiona solo la cartella Dev
corgab --path ~/Dev

# Mostra tutto ciò che è inattivo da più di 7 giorni
corgab --days 7

# Simula senza eliminare
corgab --dry-run

# Controlla quanto spazio hai liberato nel tempo
corgab --stats

# Riconfigura i target da monitorare
corgab --reset-config
```

### Controlli TUI

**Wizard di Configurazione:**

| Tasto       | Azione             |
|-------------|--------------------|
| `Spazio`    | Toggle selezione   |
| `a`         | Toggle tutti       |
| `j` / `k`  | Naviga giù/su      |
| `Enter`     | Conferma           |
| `q` / `Esc` | Esci              |

**Vista Principale:**

| Tasto       | Azione                  |
|-------------|-------------------------|
| `Spazio`    | Toggle selezione        |
| `a`         | Seleziona tutti         |
| `n`         | Deseleziona tutti       |
| `j` / `k`  | Naviga giù/su           |
| `Enter`     | Procedi all'eliminazione|
| `q` / `Esc` | Esci                   |

**Conferma:**

| Tasto | Azione          |
|-------|-----------------|
| `y`   | Conferma        |
| `n`   | Torna indietro  |

## Configurazione

Corgab salva la sua configurazione in `~/.corgab.yaml`:

```yaml
days: 30
targets:
  - node_modules
  - vendor
excluded_paths:
  - /Users/me/progetto-importante
```

- **days** — soglia di inattività predefinita (sovrascrivibile con `--days`)
- **targets** — nomi delle cartelle di dipendenze da cercare (impostati durante il wizard)
- **excluded_paths** — directory da saltare completamente durante la scansione

Modifica il file direttamente oppure usa `--reset-config` per rieseguire il wizard.

## Come Funziona

1. **Percorre** il filesystem dalla directory radice (default `~`)
2. **Salta** directory nascoste, percorsi di sistema e percorsi esclusi
3. **Trova** directory che corrispondono ai target selezionati (`node_modules`, `vendor`, ecc.)
4. **Verifica** il file di configurazione nella directory genitore (`package.json`, `composer.json`, ecc.)
5. **Filtra** — mostra solo i progetti il cui file di configurazione è più vecchio della soglia
6. **Disambigua** nomi condivisi (es. `vendor` → verifica `composer.json` vs `go.mod`)
7. **Calcola** le dimensioni in parallelo usando un pool limitato di goroutine
8. **Mostra** i risultati in una TUI interattiva per selezione ed eliminazione

Viene eliminata solo la cartella di dipendenze (es. `node_modules/`), mai il progetto stesso.

## Statistiche

Corgab tiene traccia delle statistiche cumulative in `~/.corgab_stats.json`:

```bash
$ corgab --stats
corgab cleaner has freed 14.3 GB in 47 operations since 15 Jan 2026
```

## Struttura del Progetto

```
cleaner/
├── cmd/corgab/            # Punto di ingresso, parsing dei flag
├── internal/
│   ├── config/            # Caricamento/salvataggio config YAML
│   ├── scanner/           # Scanner parallelo del filesystem
│   ├── filter/            # Rilevamento inattività (staleness)
│   ├── cleaner/           # Eliminazione directory con supporto dry-run
│   ├── stats/             # Statistiche cumulative (JSON)
│   ├── fsutil/            # Utilità condivise (DirSize, FormatBytes)
│   └── tui/               # TUI Bubbletea (wizard + vista principale)
├── pkg/
│   └── targets/           # Registro dei target
├── go.mod
└── go.sum
```

## Eseguire i Test

```bash
go test ./... -v
```

## Licenza

MIT
