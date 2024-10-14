# Implémentation du système de logging pour le Convertisseur JSON-LD

Objectif : Créer un système de logging flexible et thread-safe en Go, s'inspirant du système existant dans le projet Translator.

## Tâches :

1. Dans le répertoire `internal/logger`, créez un fichier `logger.go` avec le contenu suivant :

```go
package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

var (
	logLevel     LogLevel
	logFile      *os.File
	console      io.Writer
	mu           sync.Mutex
	silentMode   bool
	debugMode    bool
)

func Init(level LogLevel, filePath string) error {
	mu.Lock()
	defer mu.Unlock()

	logLevel = level

	if filePath != "" {
		var err error
		logFile, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
	}

	console = os.Stdout
	return nil
}

func SetLogLevel(level LogLevel) {
	mu.Lock()
	defer mu.Unlock()
	logLevel = level
}

func SetSilentMode(silent bool) {
	mu.Lock()
	defer mu.Unlock()
	silentMode = silent
}

func SetDebugMode(debug bool) {
	mu.Lock()
	defer mu.Unlock()
	debugMode = debug
	if debug {
		logLevel = DEBUG
	}
}

func log(level LogLevel, message string) {
	mu.Lock()
	defer mu.Unlock()

	if level < logLevel {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] %s: %s\n", timestamp, getLevelString(level), message)

	if logFile != nil {
		logFile.WriteString(logMessage)
	}

	if !silentMode {
		fmt.Fprint(console, logMessage)
	}
}

func getLevelString(level LogLevel) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARNING:
		return "WARNING"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

func Debug(message string) {
	if debugMode {
		log(DEBUG, message)
	}
}

func Info(message string) {
	log(INFO, message)
}

func Warning(message string) {
	log(WARNING, message)
}

func Error(message string) {
	log(ERROR, message)
}

func Close() {
	mu.Lock()
	defer mu.Unlock()

	if logFile != nil {
		logFile.Close()
	}
}
```

2. Créez un fichier de test `logger_test.go` dans le même répertoire avec le contenu suivant :

```go
package logger

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestLogging(t *testing.T) {
	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Initialize logger
	Init(INFO, "")
	SetDebugMode(true)

	// Test logging
	Debug("This is a debug message")
	Info("This is an info message")
	Warning("This is a warning message")
	Error("This is an error message")

	// Reset stdout
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Check if all messages are present
	if !strings.Contains(output, "DEBUG") {
		t.Error("Debug message not found in output")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("Info message not found in output")
	}
	if !strings.Contains(output, "WARNING") {
		t.Error("Warning message not found in output")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("Error message not found in output")
	}

	// Test silent mode
	SetSilentMode(true)
	Info("This message should not appear")

	if strings.Contains(output, "This message should not appear") {
		t.Error("Silent mode failed")
	}

	// Clean up
	Close()
}
```

3. Ajoutez les fonctionnalités suivantes au système de logging :
   - Support pour les niveaux de log : DEBUG, INFO, WARNING, ERROR
   - Écriture dans des fichiers texte et sur la console
   - Possibilité de changer le niveau de log en cours d'exécution
   - Mode silencieux (--silent) pour désactiver la sortie console
   - Mode debug (--debug) pour une journalisation très détaillée
   - Thread-safety pour une utilisation dans un environnement concurrent

4. Assurez-vous que le système de logging est facilement intégrable dans les autres parties du projet.

5. Documentez l'utilisation du système de logging dans un fichier README.md dans le répertoire `internal/logger`.

## Utilisation du système de logging :

Pour utiliser le système de logging dans d'autres parties du projet, importez le package et utilisez-le comme suit :

```go
import "github.com/votre-username/json-ld-converter/internal/logger"

func main() {
    // Initialisation du logger
    err := logger.Init(logger.INFO, "app.log")
    if err != nil {
        fmt.Printf("Erreur lors de l'initialisation du logger : %v\n", err)
        return
    }
    defer logger.Close()

    // Utilisation du logger
    logger.Debug("Message de débogage")
    logger.Info("Message d'information")
    logger.Warning("Message d'avertissement")
    logger.Error("Message d'erreur")

    // Changement du niveau de log en cours d'exécution
    logger.SetLogLevel(logger.DEBUG)

    // Activation du mode silencieux
    logger.SetSilentMode(true)

    // Activation du mode debug
    logger.SetDebugMode(true)
}
```

## Notes importantes :
- Le système de logging utilise des mutex pour assurer la thread-safety.
- Le mode debug active automatiquement le niveau de log DEBUG.
- Le mode silencieux désactive uniquement la sortie console, les logs sont toujours écrits dans le fichier si spécifié.
- N'oubliez pas d'appeler `logger.Close()` à la fin de votre programme pour fermer proprement le fichier de log.
- Les tests unitaires fournis couvrent les principales fonctionnalités du système de logging.

Veuillez implémenter ce système de logging et effectuer les tests nécessaires. Une fois terminé, nous pourrons passer à l'étape suivante du développement du convertisseur JSON-LD.