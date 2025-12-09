package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
	"github.com/Mario2334/startgen/llm"
	"github.com/briandowns/spinner"
	"github.com/spf13/pflag"
)

const (
	appName        = "startgen"
	appDescription = "Generate project boilerplate using LLMs"
)

// SupportedProviders is a list of supported LLM providers
var supportedProviders = []string{"openai", "claude", "ollama"}

// validateProvider checks if the provided provider is supported
func validateProvider(provider string) bool {
	for _, p := range supportedProviders {
		if p == provider {
			return true
		}
	}
	return false
}

// displayHelp shows the help message with usage examples
func displayHelp() {
	helpText := `%s - %s

Usage:
  %s [flags] "project description"

Flags:
  -h, --help            Show this help message
  -p, --provider string  LLM provider to use (default "openai")
  -o, --output-dir string
                        Directory to save the generated boilerplate (default ".")
  -l, --providers       Show information about available providers
  --save-metadata       Save thinking and planning outputs to .startgen-metadata.json

Examples:
  %s -p openai "A simple Go web server"
  %s --output-dir ~/projects "A Python CLI tool"
  %s --save-metadata "A REST API with authentication"
`
	fmt.Printf(helpText, appName, appDescription, appName, appName, appName, appName)
}

// displayProviderInfo shows information about the required environment variables for each provider
func displayProviderInfo() {
	const providerInfo = `
Provider-specific environment variables:
-------------------------------------
OpenAI: OPENAI_API_KEY must be set
Claude: ANTHROPIC_API_KEY must be set
Ollama: OLLAMA_ENDPOINT (optional, defaults to http://localhost:11434)
        OLLAMA_MODEL (optional, defaults to llama3)
-------------------------------------`
	fmt.Println(providerInfo)
}

// saveMetadataFile writes the thinking and planning outputs to a JSON file
func saveMetadataFile(metadata *llm.GenerationMetadata, projectPath string) error {
	metadataPath := filepath.Join(projectPath, ".startgen-metadata.json")
	
	// Create a simplified structure for the metadata file
	metadataContent := map[string]interface{}{
		"generated_at": metadata.GeneratedAt.Format(time.RFC3339),
		"thinking":     metadata.Thinking,
		"planning":     metadata.Planning,
	}
	
	// Marshal to JSON with indentation for readability
	jsonData, err := json.MarshalIndent(metadataContent, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(metadataPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write metadata file: %w", err)
	}
	
	return nil
}

// createFilesFromStructure recursively creates directories and empty files based on the provided structure
func createFilesFromStructure(structure map[string]interface{}, parentDir string) error {
	for name, value := range structure {
		currentPath := filepath.Join(parentDir, name)
		
		switch v := value.(type) {
		case map[string]interface{}:
			if err := os.MkdirAll(currentPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", currentPath, err)
			}
			if err := createFilesFromStructure(v, currentPath); err != nil {
				return err
			}
		case nil:
			if err := os.MkdirAll(filepath.Dir(currentPath), 0755); err != nil {
				return fmt.Errorf("failed to create parent directory for %s: %w", currentPath, err)
			}
			file, err := os.Create(currentPath)
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", currentPath, err)
			}
			file.Close()
		default:
			return fmt.Errorf("unexpected type %T in project structure", value)
		}
	}
	return nil
}

func main() {
	// Parse CLI flags
	outputDir := pflag.StringP("output-dir", "o", ".", "Directory to save the generated boilerplate")
	providerFlag := pflag.StringP("provider", "p", "openai", "LLM provider to use (openai, claude, ollama)")
	showProvidersFlag := pflag.BoolP("providers", "l", false, "Show information about available providers")
	helpFlag := pflag.BoolP("help", "h", false, "Show help message")
	saveMetadata := pflag.Bool("save-metadata", false, "Save thinking and planning outputs to .startgen-metadata.json")
	pflag.Parse()

	// Show help if requested
	if *helpFlag {
		displayHelp()
		return
	}

	// Show provider info if requested
	if *showProvidersFlag {
		displayProviderInfo()
		return
	}

	// Get project description from positional arguments
	args := pflag.Args()
	if len(args) < 1 {
		log.Fatal("Error: Project description is required. Use --help for usage information.")
	}
	description := args[0]

	// Validate provider
	if !validateProvider(*providerFlag) {
		log.Fatalf("Error: Unsupported provider '%s'. Supported providers: %v", 
			*providerFlag, supportedProviders)
	}

	// Initialize provider
	provider, err := llm.GetProvider(*providerFlag)
	if err != nil {
		log.Fatalf("Error initializing provider '%s': %v\nUse --providers flag for setup instructions.", 
			*providerFlag, err)
	}

	// Generate boilerplate with multi-stage process
	spin := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	spin.Start()
	defer spin.Stop()

	// Stage 1: Thinking
	spin.Suffix = fmt.Sprintf(" Analyzing requirements using %s...", *providerFlag)
	
	// Execute multi-stage generation with progress updates
	var boilerplate *llm.BoilerplateResponse
	var metadata *llm.GenerationMetadata
	
	// We'll manually orchestrate to update spinner between stages
	// Stage 1: Thinking
	thinking, err := provider.ExecuteThinkingStage(description)
	if err != nil {
		spin.Stop()
		log.Fatalf("Error in thinking stage: %v", err)
	}
	
	// Update spinner for planning stage
	spin.Suffix = fmt.Sprintf(" Creating implementation plan using %s...", *providerFlag)
	
	// Stage 2: Planning
	planning, err := provider.ExecutePlanningStage(description, thinking)
	if err != nil {
		spin.Stop()
		log.Fatalf("Error in planning stage: %v", err)
	}
	
	// Update spinner for generation stage
	spin.Suffix = fmt.Sprintf(" Generating code using %s...", *providerFlag)
	
	// Stage 3: Generation
	boilerplate, err = provider.ExecuteGenerationStage(description, thinking, planning)
	if err != nil {
		spin.Stop()
		log.Fatalf("Error in generation stage: %v", err)
	}
	
	// Create metadata
	metadata = &llm.GenerationMetadata{
		Thinking:    thinking,
		Planning:    planning,
		GeneratedAt: time.Now(),
	}

	if len(boilerplate.ProjectStructure) == 0 {
		spin.Stop()
		log.Fatal("Error: No project structure received from the LLM provider")
	}

	// Get project name from structure
	var projectName string
	for key := range boilerplate.ProjectStructure {
		projectName = key
		break
	}

	// Create project structure
	projectPath := filepath.Clean(filepath.Join(*outputDir, projectName))
	if err := createFilesFromStructure(boilerplate.ProjectStructure, *outputDir); err != nil {
		spin.Stop()
		log.Fatalf("Error creating project structure: %v", err)
	}

	// Write code files
	for filePath, content := range boilerplate.BoilerplateCode {
		fullPath := filepath.Join(*outputDir, filePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			spin.Stop()
			log.Fatalf("Error creating directory for %s: %v", fullPath, err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			spin.Stop()
			log.Fatalf("Error writing to file %s: %v", fullPath, err)
		}
	}

	spin.Stop()
	
	// Save metadata file if flag is enabled
	if *saveMetadata {
		if err := saveMetadataFile(metadata, projectPath); err != nil {
			log.Printf("Warning: Failed to save metadata file: %v", err)
		}
	}
	
	// Display success message with metadata
	fmt.Println("\n🎉 Success! Your Project Has Been Generated 🚀")
	fmt.Println("--------------------------------------------")
	fmt.Printf("🔸 Provider Used: %s\n", *providerFlag)
	fmt.Printf("🔸 Project Name: %s\n", projectName)
	fmt.Printf("🔸 Project Path: %s\n", projectPath)
	
	// Display metadata information
	if metadata != nil && metadata.Thinking != nil {
		fmt.Println("\n📋 Generation Metadata:")
		
		if len(metadata.Thinking.Technologies) > 0 {
			fmt.Println("  Technologies:")
			for _, tech := range metadata.Thinking.Technologies {
				fmt.Printf("    • %s\n", tech)
			}
		}
		
		if metadata.Thinking.ArchitecturePattern != "" {
			fmt.Printf("  Architecture Pattern: %s\n", metadata.Thinking.ArchitecturePattern)
		}
		
		if len(metadata.Thinking.KeyComponents) > 0 {
			fmt.Println("  Key Components:")
			for _, component := range metadata.Thinking.KeyComponents {
				fmt.Printf("    • %s\n", component)
			}
		}
	}
	
	fmt.Println("--------------------------------------------")
	
	if *saveMetadata {
		fmt.Printf("📄 Metadata saved to: %s\n", filepath.Join(projectPath, ".startgen-metadata.json"))
		fmt.Println("--------------------------------------------")
	}
	
	fmt.Println("You can now navigate to the project path and start building your tool. Happy coding! 💻✨")
}
