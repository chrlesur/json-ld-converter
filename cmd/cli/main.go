package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/chrlesur/json-ld-converter/internal/config"
	"github.com/chrlesur/json-ld-converter/internal/converter"
	"github.com/chrlesur/json-ld-converter/internal/llm"
	"github.com/chrlesur/json-ld-converter/internal/logger"
	"github.com/chrlesur/json-ld-converter/internal/parser"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	cfgFile      string
	inputFile    string
	outputFile   string
	engine       string
	instructions string
	silent       bool
	debug        bool
	batchMode    bool
	interactive  bool
)

var rootCmd = &cobra.Command{
	Use:               "json-ld-converter",
	Short:             "Convert documents to JSON-LD",
	Long:              `A CLI tool to convert various document formats to JSON-LD using Schema.org vocabulary.`,
	PersistentPreRunE: configureLLM,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("JSON-LD Converter v0.3.0 Alpha")
		err := convert()
		if err != nil {
			logger.Error(fmt.Sprintf("Conversion error: %v", err))
			os.Exit(1)
		}
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&inputFile, "input", "i", "", "Input file to convert")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Output file for JSON-LD")
	rootCmd.PersistentFlags().StringVarP(&engine, "engine", "e", "", "LLM engine to use (overrides config)")
	rootCmd.PersistentFlags().StringVarP(&instructions, "instructions", "n", "", "Additional instructions for LLM")
	rootCmd.PersistentFlags().BoolVar(&silent, "silent", false, "Silent mode (no console output)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Debug mode (verbose logging)")
	rootCmd.PersistentFlags().BoolVarP(&batchMode, "batch", "b", false, "Batch processing mode")
	rootCmd.PersistentFlags().BoolVarP(&interactive, "interactive", "t", false, "Interactive mode")

	// Ajout des sous-commandes
	rootCmd.AddCommand(newBatchCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newInteractiveCmd())
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		config.Load(cfgFile)
	} else {
		// Search config in home directory with name ".json-ld-converter" (without extension).
		config.Load("config.yaml")
	}

	// Configuration du logger
	logLevel := logger.INFO
	if debug {
		logLevel = logger.DEBUG
	}
	logger.Init(logLevel, "")
	logger.SetSilentMode(silent)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func convert() error {
	cfg := config.Get()

	// Création du client LLM
	client, err := llm.NewLLMClient(cfg)
	if err != nil {
		return fmt.Errorf("error creating LLM client: %w", err)
	}

	// Création du parseur de document
	p, err := parser.NewParser(getFileType(*inputFile))
	if err != nil {
		return fmt.Errorf("error creating parser: %w", err)
	}

	// Lecture et analyse du fichier d'entrée
	file, err := os.Open(*inputFile)
	if err != nil {
		return fmt.Errorf("error opening input file: %w", err)
	}
	defer file.Close()

	doc, err := p.Parse(file)
	if err != nil {
		return fmt.Errorf("error parsing document: %w", err)
	}

	// Création du convertisseur
	conv := converter.NewConverter(client, cfg.Conversion.MaxTokens, cfg.Conversion.TargetBatchSize)

	// Conversion du document en JSON-LD
	ctx := context.Background()
	jsonLD, err := conv.Convert(ctx, doc, *additionalInstructions)
	if err != nil {
		return fmt.Errorf("error converting to JSON-LD: %w", err)
	}

	// Écriture du résultat dans le fichier de sortie
	err = os.WriteFile(*outputFile, []byte(jsonLD), 0644)
	if err != nil {
		return fmt.Errorf("error writing output file: %w", err)
	}

	logger.Info("Conversion completed successfully.")
	return nil
}

func newBatchCmd() *cobra.Command {
	var inputDir string
	var outputDir string

	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Process multiple files in batch mode",
		Run: func(cmd *cobra.Command, args []string) {
			files, err := ioutil.ReadDir(inputDir)
			if err != nil {
				logger.Error(fmt.Sprintf("Error reading input directory: %v", err))
				return
			}

			for _, file := range files {
				if file.IsDir() {
					continue
				}
				inputFile = filepath.Join(inputDir, file.Name())
				outputFile = filepath.Join(outputDir, file.Name()+".jsonld")
				logger.Info(fmt.Sprintf("Processing file: %s", inputFile))
				err := convert()
				if err != nil {
					logger.Error(fmt.Sprintf("Error processing file %s: %v", inputFile, err))
				}
			}
		},
	}

	cmd.Flags().StringVarP(&inputDir, "input-dir", "d", "", "Input directory for batch processing")
	cmd.Flags().StringVarP(&outputDir, "output-dir", "o", "", "Output directory for batch processing")
	cmd.MarkFlagRequired("input-dir")
	cmd.MarkFlagRequired("output-dir")

	return cmd
}

func newConfigCmd() *cobra.Command {
	var showConfig bool
	var setKey string
	var setValue string

	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Run: func(cmd *cobra.Command, args []string) {
			if showConfig {
				cfg := config.Get()
				data, _ := yaml.Marshal(cfg)
				fmt.Println(string(data))
				return
			}

			if setKey != "" && setValue != "" {
				err := config.Set(setKey, setValue)
				if err != nil {
					logger.Error(fmt.Sprintf("Error setting config: %v", err))
					return
				}
				logger.Info(fmt.Sprintf("Config updated: %s = %s", setKey, setValue))
			}
		},
	}

	cmd.Flags().BoolVar(&showConfig, "show", false, "Show current configuration")
	cmd.Flags().StringVar(&setKey, "set-key", "", "Set configuration key")
	cmd.Flags().StringVar(&setValue, "set-value", "", "Set configuration value")

	return cmd
}

func newInteractiveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "interactive",
		Short: "Start interactive mode",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Welcome to interactive mode!")
			for {
				fmt.Print("Enter input file path (or 'quit' to exit): ")
				var input string
				fmt.Scanln(&input)
				if input == "quit" {
					break
				}
				inputFile = input
				fmt.Print("Enter output file path: ")
				fmt.Scanln(&outputFile)
				err := convert()
				if err != nil {
					logger.Error(fmt.Sprintf("Conversion error: %v", err))
				} else {
					logger.Info("Conversion completed successfully.")
				}
			}
		},
	}
}

func readAndParseDocument(filePath string, p parser.Parser) (*parser.Document, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening input file: %w", err)
	}
	defer file.Close()

	doc, err := p.Parse(file)
	if err != nil {
		return nil, fmt.Errorf("error parsing document: %w", err)
	}

	return doc, nil
}

func writeOutput(filePath string, content string) error {
	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("error writing output file: %w", err)
	}
	return nil
}

func getFileType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".txt":
		return "text"
	case ".md":
		return "markdown"
	case ".pdf":
		return "pdf"
	case ".html", ".htm":
		return "html"
	default:
		return "text" // Par défaut, on suppose que c'est un fichier texte
	}
}

func configureLLM(cmd *cobra.Command, args []string) error {
	cfg := config.Get()

	// Sélection du moteur LLM
	if engine != "" {
		cfg.Conversion.Engine = engine
	}

	// Configuration spécifique au LLM
	switch cfg.Conversion.Engine {
	case "claude":
		apiKey := os.Getenv("CLAUDE_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("CLAUDE_API_KEY environment variable is not set")
		}
		cfg.Conversion.APIKey = apiKey
	case "gpt":
		apiKey := os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return fmt.Errorf("OPENAI_API_KEY environment variable is not set")
		}
		cfg.Conversion.APIKey = apiKey
	case "ollama":
		// Pas besoin de clé API pour Ollama
		cfg.Conversion.OllamaHost = cfg.Conversion.OllamaHost
		cfg.Conversion.OllamaPort = cfg.Conversion.OllamaPort
	case "aiyou":
		email := os.Getenv("AIYOU_EMAIL")
		password := os.Getenv("AIYOU_PASSWORD")
		if email == "" || password == "" {
			return fmt.Errorf("AIYOU_EMAIL or AIYOU_PASSWORD environment variable is not set")
		}
		cfg.Conversion.AIYOUEmail = email
		cfg.Conversion.AIYOUPassword = password
	default:
		return fmt.Errorf("unsupported LLM engine: %s", cfg.Conversion.Engine)
	}

	// Application des instructions supplémentaires
	if instructions != "" {
		cfg.Conversion.AdditionalInstructions = instructions
	}

	return nil
}
