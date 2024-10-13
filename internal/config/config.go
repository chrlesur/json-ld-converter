package config

import (
	"fmt"
	"io/ioutil"
	"os"

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
		MaxTokens       int    `yaml:"max_tokens"`
		TargetBatchSize int    `yaml:"target_batch_size"`
		NumThreads      int    `yaml:"num_threads"`
		Engine          string `yaml:"engine"`
	} `yaml:"conversion"`
	Schema struct {
		Version string `yaml:"version"`
	} `yaml:"schema"`
	Segmentation struct {
		MaxTokens       int `yaml:"max_tokens"`
		TargetBatchSize int `yaml:"target_batch_size"`
	} `yaml:"segmentation"`
	Schema struct {
		FilePath string `yaml:"file_path"`
		Version  string `yaml:"version"`
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
