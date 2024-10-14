package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
	Logging struct {
		Level string `yaml:"level"`
		File  string `yaml:"file"`
	} `yaml:"logging"`
	Conversion struct {
		MaxTokens              int    `yaml:"max_tokens"`
		TargetBatchSize        int    `yaml:"target_batch_size"`
		NumThreads             int    `yaml:"num_threads"`
		Engine                 string `yaml:"engine"`
		Model                  string `yaml:"model"`
		ContextSize            int    `yaml:"context_size"`
		Timeout                int    `yaml:"timeout"`
		OllamaHost             string `yaml:"ollama_host"`
		OllamaPort             string `yaml:"ollama_port"`
		AIYOUAssistantID       string `yaml:"aiyou_assistant_id"`
		APIKey                 string `yaml:"api_key"`
		AIYOUEmail             string `yaml:"aiyou_email"`
		AIYOUPassword          string `yaml:"aiyou_password"`
		AdditionalInstructions string `yaml:"additional_instructions"`
	} `yaml:"conversion"`
	Schema struct {
		Version  string `yaml:"version"`
		FilePath string `yaml:"file_path"`
	} `yaml:"schema"`
	Segmentation struct {
		MaxTokens       int `yaml:"max_tokens"`
		TargetBatchSize int `yaml:"target_batch_size"`
	} `yaml:"segmentation"`
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
	if model := os.Getenv("CONVERSION_MODEL"); model != "" {
		c.Conversion.Model = model
	}
	if contextSize := os.Getenv("CONTEXT_SIZE"); contextSize != "" {
		fmt.Sscanf(contextSize, "%d", &c.Conversion.ContextSize)
	}
	if timeout := os.Getenv("CONVERSION_TIMEOUT"); timeout != "" {
		fmt.Sscanf(timeout, "%d", &c.Conversion.Timeout)
	}
	if ollamaHost := os.Getenv("OLLAMA_HOST"); ollamaHost != "" {
		c.Conversion.OllamaHost = ollamaHost
	}
	if ollamaPort := os.Getenv("OLLAMA_PORT"); ollamaPort != "" {
		c.Conversion.OllamaPort = ollamaPort
	}
	if aiyouAssistantID := os.Getenv("AIYOU_ASSISTANT_ID"); aiyouAssistantID != "" {
		c.Conversion.AIYOUAssistantID = aiyouAssistantID
	}
	if schemaVersion := os.Getenv("SCHEMA_VERSION"); schemaVersion != "" {
		c.Schema.Version = schemaVersion
	}
}

func Set(key, value string) error {
	// Cette implémentation est simplifiée et ne gère que les clés de premier niveau
	switch key {
	case "server.port":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid port number: %s", value)
		}
		cfg.Server.Port = port
	case "server.host":
		cfg.Server.Host = value
	// Ajoutez d'autres cas pour les autres clés de configuration
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}
	return nil
}
