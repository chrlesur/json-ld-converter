package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/chrlesur/json-ld-converter/internal/config"
	"github.com/chrlesur/json-ld-converter/internal/jsonld"
	"github.com/chrlesur/json-ld-converter/internal/llm"
	"github.com/chrlesur/json-ld-converter/internal/logger"
	"github.com/chrlesur/json-ld-converter/internal/parser"
	"github.com/chrlesur/json-ld-converter/internal/schema"
	"github.com/chrlesur/json-ld-converter/internal/segmentation"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	cfgFile      string
	inputFile    *string
	outputFile   *string
	engine       string
	instructions string
	silent       bool
	debug        bool
	batchMode    bool
	interactive  bool
)

var rootCmd = &cobra.Command{
	Use:   "json-ld-converter",
	Short: "Convert documents to JSON-LD",
	Long:  `A CLI tool to convert various document formats to JSON-LD using Schema.org vocabulary.`,
}

var convertCmd = &cobra.Command{
	Use:   "convert [input file]",
	Short: "Convert a file to JSON-LD",
	Args:  cobra.ExactArgs(1),
	RunE:  runConvert,
}

func init() {
	cobra.OnInitialize(initConfig)

	// Ajout de la sous-commande "convert"
	rootCmd.AddCommand(convertCmd)

	outputFile = new(string)
	// Flags pour la sous-commande "convert"
	convertCmd.Flags().StringVarP(outputFile, "output", "o", "", "Output file for JSON-LD (default is inputfile.jsonld)")
	convertCmd.Flags().StringVarP(&engine, "engine", "e", "", "LLM engine to use (overrides config)")
	convertCmd.Flags().StringVarP(&instructions, "instructions", "n", "", "Additional instructions for LLM")

	// Flags globaux
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&silent, "silent", false, "Silent mode (no console output)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Debug mode (verbose logging)")

	// Flags pour les autres sous-commandes (si nécessaire)
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
	logger.SetDebugMode(debug)

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func convert(inputFilePath, outputFilePath string) error {
	cfg := config.Get()
	logger.Debug(fmt.Sprintf("Configuration loaded: %+v", cfg))

	logger.Debug(fmt.Sprintf("Input file: %s", inputFilePath))
	logger.Debug(fmt.Sprintf("Output file: %s", outputFilePath))

	// Création du client LLM
	client, err := llm.NewLLMClient(cfg)
	if err != nil {
		return fmt.Errorf("error creating LLM client: %w", err)
	}
	logger.Debug("LLM client created successfully")

	// Chargement du schéma Schema.org
	schemaOrg, err := schema.LoadSchemaOrg(cfg.Schema.FilePath)
	if err != nil {
		return fmt.Errorf("error loading Schema.org: %w", err)
	}
	logger.Debug("Schema.org loaded successfully")

	// Création du parseur de document
	p, err := parser.NewParser(getFileType(inputFilePath))
	if err != nil {
		return fmt.Errorf("error creating parser: %w", err)
	}
	logger.Debug(fmt.Sprintf("Parser created for file type: %s", getFileType(inputFilePath)))

	// Lecture et analyse du fichier d'entrée
	file, err := os.Open(inputFilePath)
	if err != nil {
		return fmt.Errorf("error opening input file: %w", err)
	}
	defer file.Close()
	logger.Debug("Input file opened successfully")

	doc, err := p.Parse(file)
	if err != nil {
		return fmt.Errorf("error parsing document: %w", err)
	}
	logger.Debug("Document parsed successfully")

	// Création du convertisseur
	conv := jsonld.NewConverter(schemaOrg, client, cfg.Conversion.MaxTokens, instructions)
	logger.Debug("Converter created successfully")

	// Segmentation du document
	segments, err := segmentation.SegmentDocument(doc, cfg.Conversion.MaxTokens)
	if err != nil {
		return fmt.Errorf("error segmenting document: %w", err)
	}
	logger.Debug(fmt.Sprintf("Document segmented into %d parts", len(segments)))

	var allResults []map[string]interface{}

	// Conversion de chaque segment en JSON-LD
	for i, segment := range segments {
		logger.Debug(fmt.Sprintf("Processing segment %d of %d", i+1, len(segments)))

		segmentDoc := &parser.Document{
			Content:  segment.Content, // Utilisez segment.Content au lieu de segment directement
			Metadata: doc.Metadata,
		}

		ctx := context.Background()
		jsonLD, err := conv.Convert(ctx, segmentDoc)
		if err != nil {
			return fmt.Errorf("error converting segment %d to JSON-LD: %w", i+1, err)
		}

		allResults = append(allResults, jsonLD)
	}

	// Combinaison de tous les résultats
	combinedResult := map[string]interface{}{
		"@context": "https://schema.org",
		"@graph":   allResults,
	}

	// Sérialisation du JSON-LD combiné
	jsonString, err := json.MarshalIndent(combinedResult, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling combined JSON-LD: %w", err)
	}

	// Écriture du résultat dans le fichier de sortie
	err = os.WriteFile(outputFilePath, jsonString, 0644)
	if err != nil {
		return fmt.Errorf("error writing output file: %w", err)
	}
	logger.Debug("JSON-LD written to output file successfully")

	logger.Info("Conversion completed successfully.")
	return nil
}

func newBatchCmd() *cobra.Command {
	var inputDir, outputDir string

	cmd := &cobra.Command{
		Use:   "batch",
		Short: "Process multiple files in batch mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			files, err := ioutil.ReadDir(inputDir)
			if err != nil {
				return fmt.Errorf("error reading input directory: %w", err)
			}

			for _, file := range files {
				if file.IsDir() {
					continue
				}
				inputFilePath := filepath.Join(inputDir, file.Name())
				outputFilePath := filepath.Join(outputDir, file.Name()+".jsonld")
				logger.Info(fmt.Sprintf("Processing file: %s", inputFilePath))
				err := convert(inputFilePath, outputFilePath)
				if err != nil {
					logger.Error(fmt.Sprintf("Error processing file %s: %v", inputFilePath, err))
				}
			}
			return nil
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
	var inputDir, outputDir string

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

				// Demander le répertoire de sortie si ce n'est pas déjà fait
				if outputDir == "" {
					fmt.Print("Enter output directory: ")
					fmt.Scanln(&outputDir)
				}

				// Utiliser filepath.Base pour obtenir le nom du fichier
				fileName := filepath.Base(input)
				inputFilePath := filepath.Join(inputDir, fileName)
				outputFilePath := filepath.Join(outputDir, fileName+".jsonld")

				err := convert(inputFilePath, outputFilePath)
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

func runConvert(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	if *outputFile == "" {
		*outputFile = inputFile + ".jsonld"
	}

	logger.Debug(fmt.Sprintf("Input file: %s", inputFile))
	logger.Debug(fmt.Sprintf("Output file: %s", *outputFile))

	// Vérifiez que le fichier d'entrée existe
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputFile)
	}

	// Appel à la fonction de conversion
	err := convert(inputFile, *outputFile)
	if err != nil {
		return fmt.Errorf("conversion error: %w", err)
	}

	logger.Info("Conversion completed successfully.")
	return nil
}
