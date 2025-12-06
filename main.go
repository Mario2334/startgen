package main

import (
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

Examples:
  %s -p openai "A simple Go web server"
  %s --output-dir ~/projects "A Python CLI tool"
`
	fmt.Printf(helpText, appName, appDescription, appName, appName, appName)
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

	// Generate boilerplate with spinner
	spin := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	spin.Start()
	defer spin.Stop()
	spin.Suffix = fmt.Sprintf(" Generating boilerplate using %s...", *providerFlag)

	boilerplate, err := provider.GenerateBoilerplate(description)
	if err != nil {
		spin.Stop()
		log.Fatalf("Error generating boilerplate: %v", err)
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
	
	// Display success message
	const successMsg = `
🎉 Success! Your Project Has Been Generated 🚀
--------------------------------------------
🔸 Provider Used: %s
🔸 Project Name: %s
🔸 Project Path: %s
--------------------------------------------
You can now navigate to the project path and start building your tool. Happy coding! 💻✨
`
	fmt.Printf(successMsg, *providerFlag, projectName, projectPath)
}
