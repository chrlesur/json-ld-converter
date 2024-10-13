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
