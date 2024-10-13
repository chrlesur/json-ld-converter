# Implémentation du système de configuration pour le Convertisseur JSON-LD

Objectif : Créer un système de configuration flexible basé sur YAML pour le projet, avec support pour les surcharges par ligne de commande.

## Tâches :

1. Installez la dépendance nécessaire pour le parsing YAML :
   ```
   go get gopkg.in/yaml.v2
   ```

2. Dans le répertoire `internal/config`, créez un fichier `config.go` avec le contenu suivant :

```go
package config

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	Logging struct {
		Level string `yaml:"level"`
		File  string `yaml:"file"`
	} `yaml:"logging"`
	Conversion struct {
		MaxTokens      int    `yaml:"max_tokens"`
		TargetBatchSize int    `yaml:"target_batch_size"`
		NumThreads     int    `yaml:"num_threads"`
		Engine         string `yaml:"engine"`
	} `yaml:"conversion"`
	Schema struct {
		Version string `yaml:"version"`
	} `yaml:"schema"`
}

var (
	cfg Config
)

func Load(configPath string) error {
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("error reading config file: %v", err)
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return fmt.Errorf("error unmarshaling config: %v", err)
	}

	return nil
}

func Get() *Config {
	return &cfg
}

func (c *Config) OverrideFromEnv() {
	if port := os.Getenv("SERVER_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &c.Server.Port)
	}
	if host := os.Getenv("SERVER_HOST"); host != "" {
		c.Server.Host = host
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		c.Logging.Level = logLevel
	}
	if logFile := os.Getenv("LOG_FILE"); logFile != "" {
		c.Logging.File = logFile
	}
	if maxTokens := os.Getenv("MAX_TOKENS"); maxTokens != "" {
		fmt.Sscanf(maxTokens, "%d", &c.Conversion.MaxTokens)
	}
	if batchSize := os.Getenv("BATCH_SIZE"); batchSize != "" {
		fmt.Sscanf(batchSize, "%d", &c.Conversion.TargetBatchSize)
	}
	if numThreads := os.Getenv("NUM_THREADS"); numThreads != "" {
		fmt.Sscanf(numThreads, "%d", &c.Conversion.NumThreads)
	}
	if engine := os.Getenv("CONVERSION_ENGINE"); engine != "" {
		c.Conversion.Engine = engine
	}
	if schemaVersion := os.Getenv("SCHEMA_VERSION"); schemaVersion != "" {
		c.Schema.Version = schemaVersion
	}
}
```

3. Créez un fichier de test `config_test.go` dans le même répertoire :

```go
package config

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	content := []byte(`
server:
  port: 8080
  host: localhost
logging:
  level: info
  file: app.log
conversion:
  max_tokens: 4000
  target_batch_size: 1000
  num_threads: 4
  engine: default
schema:
  version: "1.0"
`)
	tmpfile, err := ioutil.TempFile("", "config.*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Test loading the config
	err = Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cfg := Get()

	// Check if values are correctly loaded
	if cfg.Server.Port != 8080 {
		t.Errorf("Expected server port 8080, got %d", cfg.Server.Port)
	}
	if cfg.Logging.Level != "info" {
		t.Errorf("Expected logging level info, got %s", cfg.Logging.Level)
	}
	if cfg.Conversion.MaxTokens != 4000 {
		t.Errorf("Expected max tokens 4000, got %d", cfg.Conversion.MaxTokens)
	}
	if cfg.Schema.Version != "1.0" {
		t.Errorf("Expected schema version 1.0, got %s", cfg.Schema.Version)
	}
}

func TestOverrideFromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("MAX_TOKENS", "5000")
	os.Setenv("SCHEMA_VERSION", "2.0")

	cfg := &Config{}
	cfg.OverrideFromEnv()

	// Check if values are correctly overridden
	if cfg.Server.Port != 9090 {
		t.Errorf("Expected server port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Logging.Level != "debug" {
		t.Errorf("Expected logging level debug, got %s", cfg.Logging.Level)
	}
	if cfg.Conversion.MaxTokens != 5000 {
		t.Errorf("Expected max tokens 5000, got %d", cfg.Conversion.MaxTokens)
	}
	if cfg.Schema.Version != "2.0" {
		t.Errorf("Expected schema version 2.0, got %s", cfg.Schema.Version)
	}

	// Clean up
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("MAX_TOKENS")
	os.Unsetenv("SCHEMA_VERSION")
}
```

4. Créez un fichier de configuration YAML par défaut nommé `config.yaml` à la racine du projet :

```yaml
server:
  port: 8080
  host: localhost
logging:
  level: info
  file: app.log
conversion:
  max_tokens: 4000
  target_batch_size: 1000
  num_threads: 4
  engine: default
schema:
  version: "1.0"
```

5. Dans le fichier principal de votre application (par exemple, `cmd/server/main.go` ou `cmd/cli/main.go`), ajoutez le code suivant pour charger la configuration :

```go
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/votre-username/json-ld-converter/internal/config"
)

func main() {
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	cfg := config.Get()
	cfg.OverrideFromEnv()

	fmt.Printf("Server will run on %s:%d\n", cfg.Server.Host, cfg.Server.Port)
	fmt.Printf("Logging level: %s\n", cfg.Logging.Level)
	fmt.Printf("Max tokens: %d\n", cfg.Conversion.MaxTokens)

	// Le reste de votre logique d'application ici
}
```

## Utilisation du système de configuration :

Pour utiliser le système de configuration dans d'autres parties du projet, importez le package et utilisez-le comme suit :

```go
import "github.com/votre-username/json-ld-converter/internal/config"

func someFunction() {
    cfg := config.Get()
    maxTokens := cfg.Conversion.MaxTokens
    // Utilisez la configuration comme nécessaire
}
```

## Notes importantes :
- Le système de configuration supporte le chargement à partir d'un fichier YAML.
- Les valeurs de configuration peuvent être surchargées par des variables d'environnement.
- Assurez-vous que le fichier `config.yaml` est présent à l'emplacement attendu lors de l'exécution de l'application.
- Pour les surcharges par ligne de commande, vous devrez implémenter la logique dans votre CLI en utilisant un package comme `flag` ou `cobra`.
- Les tests unitaires fournis couvrent le chargement de la configuration et les surcharges par variables d'environnement.

Veuillez implémenter ce système de configuration et effectuer les tests nécessaires. Une fois terminé, nous pourrons passer à l'étape suivante du développement du convertisseur JSON-LD.