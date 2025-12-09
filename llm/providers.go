package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

const (
	defaultTimeout = 120 * time.Second
)

// Provider defines the interface that all LLM providers must implement.
type Provider interface {
	// ExecuteThinkingStage executes the thinking stage of multi-stage generation.
	// It analyzes the project description and returns a ThinkingOutput containing
	// identified technologies, architecture patterns, key components, assumptions, and constraints.
	ExecuteThinkingStage(description string) (*ThinkingOutput, error)
	
	// ExecutePlanningStage executes the planning stage of multi-stage generation.
	// It takes the project description and thinking output as context, and returns
	// a PlanningOutput containing the project structure, file descriptions, dependencies,
	// and configuration files.
	ExecutePlanningStage(description string, thinking *ThinkingOutput) (*PlanningOutput, error)
	
	// ExecuteGenerationStage executes the generation stage of multi-stage generation.
	// It takes the project description, thinking output, and planning output as context,
	// and returns a BoilerplateResponse containing the actual code for all planned files.
	ExecuteGenerationStage(description string, thinking *ThinkingOutput, planning *PlanningOutput) (*BoilerplateResponse, error)
	
	// GenerateBoilerplate generates project boilerplate based on the given description.
	// It returns a BoilerplateResponse containing the project structure and boilerplate code,
	// or an error if the generation fails.
	// This method is maintained for backward compatibility.
	GenerateBoilerplate(description string) (*BoilerplateResponse, error)
}

// BoilerplateResponse represents the structured response from an LLM.
type BoilerplateResponse struct {
	// ProjectStructure defines the directory and file structure of the project.
	// The keys are file/directory names, and the values are either nested maps
	// (for directories) or nil (for files).
	ProjectStructure map[string]interface{} `json:"project_structure"`
	
	// BoilerplateCode contains the actual code content for the generated files.
	// The keys are file paths relative to the project root, and the values
	// are the file contents as strings.
	BoilerplateCode map[string]string `json:"boilerplate_code"`
}

// ThinkingOutput represents the output from the thinking stage where the LLM
// analyzes project requirements and identifies key technologies, patterns, and constraints.
type ThinkingOutput struct {
	// Technologies lists the key technologies identified for the project
	Technologies []string `json:"technologies"`
	
	// ArchitecturePattern describes the recommended architectural pattern
	// (e.g., MVC, microservices, layered architecture)
	ArchitecturePattern string `json:"architecture_pattern"`
	
	// KeyComponents lists the main components or modules of the system
	KeyComponents []string `json:"key_components"`
	
	// Assumptions documents any assumptions made when requirements were ambiguous
	Assumptions []string `json:"assumptions"`
	
	// Constraints lists technical or business constraints identified
	Constraints []string `json:"constraints"`
}

// Dependency represents a project dependency with its purpose and optional version.
type Dependency struct {
	// Name is the dependency package or library name
	Name string `json:"name"`
	
	// Purpose explains why this dependency is needed
	Purpose string `json:"purpose"`
	
	// Version specifies the dependency version (optional)
	Version string `json:"version,omitempty"`
}

// ConfigFile represents a configuration file in the project plan.
type ConfigFile struct {
	// Path is the file path relative to the project root
	Path string `json:"path"`
	
	// Format specifies the configuration file format (e.g., JSON, YAML, TOML)
	Format string `json:"format"`
	
	// Purpose explains what this configuration file is used for
	Purpose string `json:"purpose"`
}

// PlanningOutput represents the output from the planning stage where the LLM
// creates a structured implementation plan with directory structure and file descriptions.
type PlanningOutput struct {
	// ProjectName is the name of the project
	ProjectName string `json:"project_name"`
	
	// DirectoryStructure defines the hierarchical directory and file structure
	// The keys are file/directory names, and the values are either nested maps
	// (for directories) or nil (for files)
	DirectoryStructure map[string]interface{} `json:"directory_structure"`
	
	// FileDescriptions maps file paths to descriptions of their purpose
	FileDescriptions map[string]string `json:"file_descriptions"`
	
	// Dependencies lists all required dependencies for the project
	Dependencies []Dependency `json:"dependencies"`
	
	// ConfigurationFiles lists all configuration files that will be created
	ConfigurationFiles []ConfigFile `json:"configuration_files"`
}

// StageType represents the different stages of multi-stage generation.
type StageType int

const (
	// ThinkingStage is the initial reasoning phase where the LLM analyzes project requirements
	ThinkingStage StageType = iota
	// PlanningStage is the intermediate phase where the LLM creates a structured implementation plan
	PlanningStage
	// GenerationStage is the final phase where actual code and file structure are created
	GenerationStage
)

// PromptBuilder constructs stage-specific prompts for multi-stage generation.
type PromptBuilder struct{}

// NewPromptBuilder creates a new PromptBuilder instance.
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

// BuildThinkingPrompt creates a prompt for the thinking stage that analyzes project requirements.
// It instructs the LLM to identify technologies, architecture patterns, key components,
// assumptions, and constraints from the project description.
func (pb *PromptBuilder) BuildThinkingPrompt(description string) string {
	return fmt.Sprintf(`You are a software architect analyzing project requirements. Your task is to analyze the following project description and provide a structured analysis.

Project Description:
%s

Please analyze this project and provide your response in the following JSON format:

{
  "technologies": ["list", "of", "technologies"],
  "architecture_pattern": "recommended architectural pattern",
  "key_components": ["list", "of", "main", "components"],
  "assumptions": ["list", "of", "assumptions", "made"],
  "constraints": ["list", "of", "identified", "constraints"]
}

Field Descriptions:
- technologies: Key technologies, frameworks, and languages that should be used (e.g., "Go", "PostgreSQL", "React")
- architecture_pattern: The recommended architectural pattern (e.g., "MVC", "microservices", "layered architecture", "REST API")
- key_components: Main components or modules of the system (e.g., "authentication service", "database layer", "API endpoints")
- assumptions: Any assumptions you're making when requirements are ambiguous (e.g., "Assuming REST API over GraphQL", "Assuming PostgreSQL for data persistence")
- constraints: Technical or business constraints identified (e.g., "Must be stateless", "Should support concurrent users", "Limited to standard library")

Example Output:
{
  "technologies": ["Python", "Flask", "SQLite", "pytest"],
  "architecture_pattern": "MVC with REST API",
  "key_components": ["Web server", "Database models", "API routes", "Authentication middleware"],
  "assumptions": ["Using Flask for simplicity", "SQLite for development database", "Token-based authentication"],
  "constraints": ["Must be lightweight", "Should have minimal dependencies", "Must support CRUD operations"]
}

Important: Return ONLY the JSON object wrapped in a code block. Do not include any additional text or explanations outside the JSON.`, description)
}

// BuildPlanningPrompt creates a prompt for the planning stage that creates a structured implementation plan.
// It includes the thinking output as context and requests a detailed plan with directory structure,
// file descriptions, dependencies, and configuration files.
func (pb *PromptBuilder) BuildPlanningPrompt(description string, thinking *ThinkingOutput) string {
	thinkingJSON, _ := json.MarshalIndent(thinking, "", "  ")
	
	return fmt.Sprintf(`You are a software architect creating an implementation plan. Based on the project description and the analysis from the thinking stage, create a detailed implementation plan.

Project Description:
%s

Analysis from Thinking Stage:
%s

Please create a structured implementation plan in the following JSON format:

{
  "project_name": "name_of_the_project",
  "directory_structure": {
    "project_name": {
      "src": {
        "main.py": null,
        "utils.py": null
      },
      "tests": {
        "test_main.py": null
      },
      "README.md": null
    }
  },
  "file_descriptions": {
    "project_name/src/main.py": "Entry point of the application",
    "project_name/src/utils.py": "Utility functions",
    "project_name/tests/test_main.py": "Unit tests for main module",
    "project_name/README.md": "Project documentation"
  },
  "dependencies": [
    {
      "name": "flask",
      "purpose": "Web framework for building the API",
      "version": "2.3.0"
    }
  ],
  "configuration_files": [
    {
      "path": "project_name/config.yaml",
      "format": "YAML",
      "purpose": "Application configuration settings"
    }
  ]
}

Field Descriptions:
- project_name: A kebab-case or snake_case name for the project
- directory_structure: A nested object representing the file system hierarchy. Directories are objects with nested content, files have null values
- file_descriptions: A map of file paths to descriptions explaining each file's purpose
- dependencies: List of required packages/libraries with their purpose and optional version
- configuration_files: List of configuration files that will be created, with their format and purpose

Important Guidelines:
1. The directory_structure must use the project_name as the root directory
2. All file paths in file_descriptions must match the structure in directory_structure
3. Include all necessary files for a complete, working project
4. Consider the technologies and architecture pattern from the thinking stage
5. Return ONLY the JSON object wrapped in a code block

Example for a Go web server:
{
  "project_name": "go-web-server",
  "directory_structure": {
    "go-web-server": {
      "cmd": {
        "server": {
          "main.go": null
        }
      },
      "internal": {
        "handlers": {
          "handlers.go": null
        },
        "models": {
          "models.go": null
        }
      },
      "go.mod": null,
      "README.md": null
    }
  },
  "file_descriptions": {
    "go-web-server/cmd/server/main.go": "Application entry point and server setup",
    "go-web-server/internal/handlers/handlers.go": "HTTP request handlers",
    "go-web-server/internal/models/models.go": "Data models and business logic",
    "go-web-server/go.mod": "Go module definition",
    "go-web-server/README.md": "Project documentation and setup instructions"
  },
  "dependencies": [
    {
      "name": "github.com/gorilla/mux",
      "purpose": "HTTP router and URL matcher",
      "version": "v1.8.0"
    }
  ],
  "configuration_files": [
    {
      "path": "go-web-server/config.json",
      "format": "JSON",
      "purpose": "Server configuration (port, timeouts, etc.)"
    }
  ]
}`, description, string(thinkingJSON))
}

// BuildGenerationPrompt creates a prompt for the generation stage that produces actual code.
// It includes both thinking and planning outputs as context and requests complete,
// working code for all files in the plan.
func (pb *PromptBuilder) BuildGenerationPrompt(description string, thinking *ThinkingOutput, planning *PlanningOutput) string {
	thinkingJSON, _ := json.MarshalIndent(thinking, "", "  ")
	planningJSON, _ := json.MarshalIndent(planning, "", "  ")
	
	return fmt.Sprintf(`You are a software developer implementing a project based on a detailed plan. Generate MINIMAL but functional boilerplate code for all files in the implementation plan.

IMPORTANT: Keep your response concise! For complex projects, provide minimal implementations with TODO comments rather than full features. The goal is to create a working skeleton, not a complete application.

Project Description:
%s

Analysis from Thinking Stage:
%s

Implementation Plan from Planning Stage:
%s

Please generate the complete project code in the following JSON format:

{
  "project_structure": {
    "project_name": {
      "src": {
        "main.py": null,
        "utils.py": null
      },
      "tests": {
        "test_main.py": null
      },
      "README.md": null
    }
  },
  "boilerplate_code": {
    "project_name/src/main.py": "def main():\n    print('Hello, World!')\n\nif __name__ == '__main__':\n    main()",
    "project_name/src/utils.py": "def helper_function():\n    pass",
    "project_name/tests/test_main.py": "import unittest\n\nclass TestMain(unittest.TestCase):\n    def test_example(self):\n        self.assertTrue(True)",
    "project_name/README.md": "# Project Name\n\nDescription of the project."
  }
}

Field Descriptions:
- project_structure: Must exactly match the directory_structure from the planning stage
- boilerplate_code: A map where keys are file paths (matching the planning stage) and values are the complete file contents

Important Guidelines:
1. Generate code for EVERY file listed in the planning stage's directory_structure
2. The project_structure must exactly match the directory_structure from planning
3. All file paths in boilerplate_code must correspond to files in the project_structure
4. Keep code CONCISE - provide minimal but functional implementations
5. For large projects, prioritize core functionality over comprehensive features
6. Use TODO comments for complex features that would make the response too long
7. Include essential imports and basic structure
8. For configuration files, generate minimal but valid content
9. Return ONLY the JSON object wrapped in a code block
10. CRITICAL: Keep the total response under 15000 tokens - be concise!

Example for a simple Python CLI tool:
{
  "project_structure": {
    "python-cli-tool": {
      "src": {
        "cli.py": null,
        "commands.py": null
      },
      "tests": {
        "test_commands.py": null
      },
      "requirements.txt": null,
      "README.md": null
    }
  },
  "boilerplate_code": {
    "python-cli-tool/src/cli.py": "#!/usr/bin/env python3\nimport argparse\nfrom commands import execute_command\n\ndef main():\n    parser = argparse.ArgumentParser(description='CLI Tool')\n    parser.add_argument('command', help='Command to execute')\n    args = parser.parse_args()\n    execute_command(args.command)\n\nif __name__ == '__main__':\n    main()",
    "python-cli-tool/src/commands.py": "def execute_command(cmd):\n    print(f'Executing: {cmd}')",
    "python-cli-tool/tests/test_commands.py": "import unittest\nfrom src.commands import execute_command\n\nclass TestCommands(unittest.TestCase):\n    def test_execute(self):\n        # Test implementation\n        pass",
    "python-cli-tool/requirements.txt": "# Add your dependencies here",
    "python-cli-tool/README.md": "# Python CLI Tool\\n\\nA simple command-line interface tool.\\n\\n## Installation\\n\\npip install -r requirements.txt\\n\\n## Usage\\n\\npython src/cli.py <command>"
  }
}`, description, string(thinkingJSON), string(planningJSON))
}

// getJSONFromMessage extracts JSON content from a message that may contain markdown code blocks.
func getJSONFromMessage(data string) (string, error) {
	// First try to find a JSON code block with language specifier
	jsonPattern := regexp.MustCompile("(?s)```json\\s*(.*?)```")
	matches := jsonPattern.FindStringSubmatch(data)

	// If no match, try to find any code block
	if len(matches) < 2 {
		codeBlockPattern := regexp.MustCompile("(?s)```(.*?)```")
		matches = codeBlockPattern.FindStringSubmatch(data)
	}

	if len(matches) < 2 {
		// If no code block found, try to extract JSON directly from the response
		// Look for content between { and } (greedy match to get the full JSON)
		directJSONPattern := regexp.MustCompile("(?s)\\{.*\\}")
		directMatches := directJSONPattern.FindString(data)
		if directMatches != "" {
			extracted := strings.TrimSpace(directMatches)
			// Validate JSON syntax
			var js map[string]interface{}
			if err := json.Unmarshal([]byte(extracted), &js); err != nil {
				return "", fmt.Errorf("invalid JSON in response: %w\nResponse preview: %s", err, truncateString(data, 500))
			}
			return extracted, nil
		}
		return "", fmt.Errorf("no code block found in the response\nResponse preview: %s", truncateString(data, 500))
	}

	extracted := strings.TrimSpace(matches[1])
	if extracted == "" {
		return "", fmt.Errorf("empty content in code block\nResponse preview: %s", truncateString(data, 500))
	}

	// Validate JSON syntax
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(extracted), &js); err != nil {
		return "", fmt.Errorf("invalid JSON in response: %w\nExtracted content preview: %s", err, truncateString(extracted, 500))
	}

	return extracted, nil
}

// truncateString truncates a string to maxLen characters and adds "..." if truncated
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// OpenAIMessage represents a message in the OpenAI chat format.
type OpenAIMessage struct {
	Role    string `json:"role"`    // "system", "user", or "assistant"
	Content string `json:"content"` // The actual message content
}

// OpenAICompletionRequest represents a request to the OpenAI chat completion API.
type OpenAICompletionRequest struct {
	Model     string          `json:"model"`
	Messages  []OpenAIMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens,omitempty"`
}

// OpenAIChoice represents a choice in the OpenAI API response.
type OpenAIChoice struct {
	Index        int           `json:"index"`
	Message      OpenAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason,omitempty"`
}

// OpenAIResponse represents the response from the OpenAI API.
type OpenAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
}

// OpenAIProvider implements the Provider interface for OpenAI's API.
type OpenAIProvider struct {
	apiKey string
	model  string
	client *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider instance.
// It reads the API key from the OPENAI_API_KEY environment variable.
func NewOpenAIProvider() (*OpenAIProvider, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, errors.New("OPENAI_API_KEY environment variable is not set")
	}

	return &OpenAIProvider{
		apiKey: apiKey,
		model:  "gpt-4o",
		client: &http.Client{Timeout: defaultTimeout},
	}, nil
}

// ExecuteThinkingStage executes the thinking stage using OpenAI's API.
// It analyzes the project description and returns a ThinkingOutput.
func (p *OpenAIProvider) ExecuteThinkingStage(description string) (*ThinkingOutput, error) {
	promptBuilder := NewPromptBuilder()
	prompt := promptBuilder.BuildThinkingPrompt(description)

	// Create the request payload
	reqBody := OpenAICompletionRequest{
		Model: p.model,
		Messages: []OpenAIMessage{
			{Role: "user", Content: prompt},
		},
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("thinking stage: failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		"https://api.openai.com/v1/chat/completions",
		bytes.NewBuffer(reqBodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("thinking stage: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Send the request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("thinking stage: API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("thinking stage: failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("thinking stage: API request failed with status %d: %s",
			resp.StatusCode, string(respBody))
	}

	// Parse the response
	var result OpenAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("thinking stage: failed to parse API response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, errors.New("thinking stage: no choices in API response")
	}

	// Extract and parse JSON from the message content
	message := result.Choices[0].Message.Content
	extractedJSON, err := getJSONFromMessage(message)
	if err != nil {
		return nil, fmt.Errorf("thinking stage: failed to extract JSON from response: %w", err)
	}

	// Parse the thinking output
	var thinking ThinkingOutput
	if err := json.Unmarshal([]byte(extractedJSON), &thinking); err != nil {
		return nil, fmt.Errorf("thinking stage: failed to parse thinking JSON: %w", err)
	}

	// Validate the response
	if len(thinking.Technologies) == 0 {
		return nil, errors.New("thinking stage: no technologies identified in response")
	}

	if thinking.ArchitecturePattern == "" {
		return nil, errors.New("thinking stage: no architecture pattern identified in response")
	}

	if len(thinking.KeyComponents) == 0 {
		return nil, errors.New("thinking stage: no key components identified in response")
	}

	return &thinking, nil
}

// ExecutePlanningStage executes the planning stage using OpenAI's API.
// It takes the thinking output as context and returns a PlanningOutput.
func (p *OpenAIProvider) ExecutePlanningStage(description string, thinking *ThinkingOutput) (*PlanningOutput, error) {
	if thinking == nil {
		return nil, errors.New("planning stage: thinking output is required")
	}

	promptBuilder := NewPromptBuilder()
	prompt := promptBuilder.BuildPlanningPrompt(description, thinking)

	// Create the request payload
	reqBody := OpenAICompletionRequest{
		Model: p.model,
		Messages: []OpenAIMessage{
			{Role: "user", Content: prompt},
		},
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("planning stage: failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		"https://api.openai.com/v1/chat/completions",
		bytes.NewBuffer(reqBodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("planning stage: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Send the request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("planning stage: API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("planning stage: failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("planning stage: API request failed with status %d: %s",
			resp.StatusCode, string(respBody))
	}

	// Parse the response
	var result OpenAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("planning stage: failed to parse API response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, errors.New("planning stage: no choices in API response")
	}

	// Extract and parse JSON from the message content
	message := result.Choices[0].Message.Content
	extractedJSON, err := getJSONFromMessage(message)
	if err != nil {
		return nil, fmt.Errorf("planning stage: failed to extract JSON from response: %w", err)
	}

	// Parse the planning output
	var planning PlanningOutput
	if err := json.Unmarshal([]byte(extractedJSON), &planning); err != nil {
		return nil, fmt.Errorf("planning stage: failed to parse planning JSON: %w", err)
	}

	// Validate the response
	if planning.ProjectName == "" {
		return nil, errors.New("planning stage: no project name in response")
	}

	if len(planning.DirectoryStructure) == 0 {
		return nil, errors.New("planning stage: empty directory structure in response")
	}

	return &planning, nil
}

// ExecuteGenerationStage executes the generation stage using OpenAI's API.
// It takes both thinking and planning outputs as context and returns a BoilerplateResponse.
func (p *OpenAIProvider) ExecuteGenerationStage(description string, thinking *ThinkingOutput, planning *PlanningOutput) (*BoilerplateResponse, error) {
	if thinking == nil {
		return nil, errors.New("generation stage: thinking output is required")
	}

	if planning == nil {
		return nil, errors.New("generation stage: planning output is required")
	}

	promptBuilder := NewPromptBuilder()
	prompt := promptBuilder.BuildGenerationPrompt(description, thinking, planning)

	// Create the request payload
	reqBody := OpenAICompletionRequest{
		Model:     p.model,
		MaxTokens: 16000, // gpt-4o supports up to 16k output tokens
		Messages: []OpenAIMessage{
			{Role: "user", Content: prompt},
		},
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("generation stage: failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		"https://api.openai.com/v1/chat/completions",
		bytes.NewBuffer(reqBodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("generation stage: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Send the request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("generation stage: API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("generation stage: failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("generation stage: API request failed with status %d: %s",
			resp.StatusCode, string(respBody))
	}

	// Parse the response
	var result OpenAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("generation stage: failed to parse API response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, errors.New("generation stage: no choices in API response")
	}

	// Check if the response was truncated due to token limits
	choice := result.Choices[0]
	
	// Always log finish reason for debugging
	fmt.Fprintf(os.Stderr, "Debug: finish_reason=%s, message_length=%d\n", choice.FinishReason, len(choice.Message.Content))
	
	if choice.FinishReason == "length" {
		return nil, fmt.Errorf("generation stage: response was truncated due to token limit (finish_reason=length).\n\nSuggestions:\n1. Simplify your project description\n2. Use more specific, focused requirements\n3. Generate components separately\n4. Avoid requesting full-stack applications in one go")
	}

	// Extract and parse JSON from the message content
	message := choice.Message.Content
	
	// Save the full response for debugging
	debugFile := "/tmp/startgen_debug_response.txt"
	if err := os.WriteFile(debugFile, []byte(message), 0644); err == nil {
		fmt.Fprintf(os.Stderr, "Debug: Full response saved to %s\n", debugFile)
	}
	
	extractedJSON, err := getJSONFromMessage(message)
	if err != nil {
		return nil, fmt.Errorf("generation stage: failed to extract JSON from response: %w", err)
	}

	// Parse the boilerplate response
	var boilerplate BoilerplateResponse
	if err := json.Unmarshal([]byte(extractedJSON), &boilerplate); err != nil {
		return nil, fmt.Errorf("generation stage: failed to parse boilerplate JSON: %w", err)
	}

	// Validate the response
	if len(boilerplate.ProjectStructure) == 0 {
		return nil, errors.New("generation stage: empty project structure in response")
	}

	if len(boilerplate.BoilerplateCode) == 0 {
		return nil, errors.New("generation stage: no boilerplate code in response")
	}

	return &boilerplate, nil
}

// GenerateBoilerplate generates project boilerplate using OpenAI's API.
func (p *OpenAIProvider) GenerateBoilerplate(description string) (*BoilerplateResponse, error) {
	// Prepare the system message with clear instructions
	systemMessage := `You are a helpful assistant that generates project boilerplate code. 
Return a JSON object with two fields:
1. 'project_structure': A nested object representing the file/directory structure.
   - Directories should be objects with nested files/directories as values.
   - Files should have a value of null.
2. 'boilerplate_code': A map where keys are file paths and values are the file contents.

Example:
{
  "project_structure": {
    "project_name": {
      "src": {
        "main.py": null,
        "utils.py": null
      },
      "tests": {
        "test_main.py": null
      },
      "README.md": null
    }
  },
  "boilerplate_code": {
    "project_name/README.md": "# Project Name\n\nA brief description of the project.",
    "project_name/src/main.py": "def main():\n    print(\"Hello, World!\")\n\nif __name__ == \"__main__\":\n    main()"
  }
}`

	// Create the request payload
	reqBody := OpenAICompletionRequest{
		Model: p.model,
		Messages: []OpenAIMessage{
			{Role: "system", Content: systemMessage},
			{
				Role:    "user",
				Content: fmt.Sprintf("Generate boilerplate for a project with this description: %s", description),
			},
		},
	}

	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		"https://api.openai.com/v1/chat/completions",
		bytes.NewBuffer(reqBodyBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	// Send the request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", 
			resp.StatusCode, string(respBody))
	}

	// Parse the response
	var result OpenAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, errors.New("no choices in API response")
	}

	// Extract and parse JSON from the message content
	message := result.Choices[0].Message.Content
	extractedJSON, err := getJSONFromMessage(message)
	if err != nil {
		return nil, fmt.Errorf("failed to extract JSON from response: %w", err)
	}

	// Parse the boilerplate response
	var boilerplate BoilerplateResponse
	if err := json.Unmarshal([]byte(extractedJSON), &boilerplate); err != nil {
		return nil, fmt.Errorf("failed to parse boilerplate JSON: %w", err)
	}

	// Validate the response
	if len(boilerplate.ProjectStructure) == 0 {
		return nil, errors.New("empty project structure in response")
	}

	if len(boilerplate.BoilerplateCode) == 0 {
		return nil, errors.New("no boilerplate code in response")
	}

	return &boilerplate, nil
}

// ClaudeProvider implements the Provider interface for Anthropic's Claude API.
type ClaudeProvider struct {
	apiKey string
	model  string
	client *http.Client
}

// Claude message structure
type ClaudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Claude API response structure
type ClaudeResponse struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
}

// NewClaudeProvider creates a new Claude provider instance.
// It reads the API key from the ANTHROPIC_API_KEY environment variable.
func NewClaudeProvider() (*ClaudeProvider, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, errors.New("ANTHROPIC_API_KEY environment variable is not set")
	}

	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = "claude-3-opus-20240229"
	}

	return &ClaudeProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: defaultTimeout},
	}, nil
}

// ExecuteThinkingStage executes the thinking stage using Claude's API.
// It analyzes the project description and returns a ThinkingOutput.
func (p *ClaudeProvider) ExecuteThinkingStage(description string) (*ThinkingOutput, error) {
	promptBuilder := NewPromptBuilder()
	prompt := promptBuilder.BuildThinkingPrompt(description)

	// Create the request payload for Claude API
	payload := map[string]interface{}{
		"model":      p.model,
		"max_tokens": 8000,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("thinking stage: failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		"https://api.anthropic.com/v1/messages",
		bytes.NewBuffer(payloadBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("thinking stage: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send the request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("thinking stage: API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("thinking stage: failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("thinking stage: API request failed with status %d: %s",
			resp.StatusCode, string(respBody))
	}

	// Parse the response
	var claudeResp ClaudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return nil, fmt.Errorf("thinking stage: failed to parse API response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, errors.New("thinking stage: no content in API response")
	}

	// Extract and parse JSON from the message content
	content := claudeResp.Content[0].Text
	extractedJSON, err := getJSONFromMessage(content)
	if err != nil {
		return nil, fmt.Errorf("thinking stage: failed to extract JSON from response: %w", err)
	}

	// Parse the thinking output
	var thinking ThinkingOutput
	if err := json.Unmarshal([]byte(extractedJSON), &thinking); err != nil {
		return nil, fmt.Errorf("thinking stage: failed to parse thinking JSON: %w", err)
	}

	// Validate the response
	if len(thinking.Technologies) == 0 {
		return nil, errors.New("thinking stage: no technologies identified in response")
	}

	if thinking.ArchitecturePattern == "" {
		return nil, errors.New("thinking stage: no architecture pattern identified in response")
	}

	if len(thinking.KeyComponents) == 0 {
		return nil, errors.New("thinking stage: no key components identified in response")
	}

	return &thinking, nil
}

// ExecutePlanningStage executes the planning stage using Claude's API.
// It takes the thinking output as context and returns a PlanningOutput.
func (p *ClaudeProvider) ExecutePlanningStage(description string, thinking *ThinkingOutput) (*PlanningOutput, error) {
	if thinking == nil {
		return nil, errors.New("planning stage: thinking output is required")
	}

	promptBuilder := NewPromptBuilder()
	prompt := promptBuilder.BuildPlanningPrompt(description, thinking)

	// Create the request payload for Claude API
	payload := map[string]interface{}{
		"model":      p.model,
		"max_tokens": 8000,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("planning stage: failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		"https://api.anthropic.com/v1/messages",
		bytes.NewBuffer(payloadBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("planning stage: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send the request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("planning stage: API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("planning stage: failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("planning stage: API request failed with status %d: %s",
			resp.StatusCode, string(respBody))
	}

	// Parse the response
	var claudeResp ClaudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return nil, fmt.Errorf("planning stage: failed to parse API response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, errors.New("planning stage: no content in API response")
	}

	// Extract and parse JSON from the message content
	content := claudeResp.Content[0].Text
	extractedJSON, err := getJSONFromMessage(content)
	if err != nil {
		return nil, fmt.Errorf("planning stage: failed to extract JSON from response: %w", err)
	}

	// Parse the planning output
	var planning PlanningOutput
	if err := json.Unmarshal([]byte(extractedJSON), &planning); err != nil {
		return nil, fmt.Errorf("planning stage: failed to parse planning JSON: %w", err)
	}

	// Validate the response
	if planning.ProjectName == "" {
		return nil, errors.New("planning stage: no project name in response")
	}

	if len(planning.DirectoryStructure) == 0 {
		return nil, errors.New("planning stage: empty directory structure in response")
	}

	return &planning, nil
}

// ExecuteGenerationStage executes the generation stage using Claude's API.
// It takes both thinking and planning outputs as context and returns a BoilerplateResponse.
func (p *ClaudeProvider) ExecuteGenerationStage(description string, thinking *ThinkingOutput, planning *PlanningOutput) (*BoilerplateResponse, error) {
	if thinking == nil {
		return nil, errors.New("generation stage: thinking output is required")
	}

	if planning == nil {
		return nil, errors.New("generation stage: planning output is required")
	}

	promptBuilder := NewPromptBuilder()
	prompt := promptBuilder.BuildGenerationPrompt(description, thinking, planning)

	// Create the request payload for Claude API
	payload := map[string]interface{}{
		"model":      p.model,
		"max_tokens": 8000,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("generation stage: failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(
		"POST",
		"https://api.anthropic.com/v1/messages",
		bytes.NewBuffer(payloadBytes),
	)
	if err != nil {
		return nil, fmt.Errorf("generation stage: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	// Send the request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("generation stage: API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("generation stage: failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("generation stage: API request failed with status %d: %s",
			resp.StatusCode, string(respBody))
	}

	// Parse the response
	var claudeResp ClaudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return nil, fmt.Errorf("generation stage: failed to parse API response: %w", err)
	}

	if len(claudeResp.Content) == 0 {
		return nil, errors.New("generation stage: no content in API response")
	}

	// Extract and parse JSON from the message content
	content := claudeResp.Content[0].Text
	extractedJSON, err := getJSONFromMessage(content)
	if err != nil {
		return nil, fmt.Errorf("generation stage: failed to extract JSON from response: %w", err)
	}

	// Parse the boilerplate response
	var boilerplate BoilerplateResponse
	if err := json.Unmarshal([]byte(extractedJSON), &boilerplate); err != nil {
		return nil, fmt.Errorf("generation stage: failed to parse boilerplate JSON: %w", err)
	}

	// Validate the response
	if len(boilerplate.ProjectStructure) == 0 {
		return nil, errors.New("generation stage: empty project structure in response")
	}

	if len(boilerplate.BoilerplateCode) == 0 {
		return nil, errors.New("generation stage: no boilerplate code in response")
	}

	return &boilerplate, nil
}

// GenerateBoilerplate implements Provider interface for Claude
func (p *ClaudeProvider) GenerateBoilerplate(description string) (*BoilerplateResponse, error) {
	url := "https://api.anthropic.com/v1/messages"
	
	// Structured payload for Claude API
	payload := map[string]interface{}{
		"model": p.model,
		"max_tokens": 4000,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": fmt.Sprintf(`Return as JSON. 
				Include: project_structure, boilerplate_code
				You are a software assistant that generates boilerplate code. 
				The sections (headers should be same) should include Project Structure (without comments), Boilerplate Code.
				The Project Structure should be represented as a nested json as a file system hierarchy and the file should have value as None
				Boilerplate Code file should directory mention from project path and it should be plain key-value pairs where key is file path and value is code text. 
				Validate the json and correct it.
				
				Generate boilerplate for a project described as: %s`, description),
			},
		},
	}
	payloadBytes, _ := json.Marshal(payload)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", 
			resp.StatusCode, string(body))
	}

	var claudeResp ClaudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	var content string
	if len(claudeResp.Content) > 0 {
		content = claudeResp.Content[0].Text
	} else {
		return nil, errors.New("no content in response")
	}
	
	extractedJSON, err := getJSONFromMessage(content)
	if err != nil {
		return nil, fmt.Errorf("error extracting JSON from message: %w", err)
	}

	var boilerplate BoilerplateResponse
	if err := json.Unmarshal([]byte(extractedJSON), &boilerplate); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return &boilerplate, nil
}

// OllamaProvider implements the Provider interface for Ollama's API.
type OllamaProvider struct {
	endpoint string
	model    string
	client   *http.Client
}

// Ollama API response structure
type OllamaResponse struct {
	Response string `json:"response"`
}

// NewOllamaProvider creates a new Ollama provider instance.
// It reads the endpoint and model from environment variables or uses defaults.
func NewOllamaProvider() (*OllamaProvider, error) {
	endpoint := os.Getenv("OLLAMA_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}

	model := os.Getenv("OLLAMA_MODEL")
	if model == "" {
		model = "llama3"
	}

	return &OllamaProvider{
		endpoint: endpoint,
		model:    model,
		client:   &http.Client{Timeout: defaultTimeout},
	}, nil
}

// ExecuteThinkingStage executes the thinking stage using Ollama's API.
// It analyzes the project description and returns a ThinkingOutput.
func (p *OllamaProvider) ExecuteThinkingStage(description string) (*ThinkingOutput, error) {
	promptBuilder := NewPromptBuilder()
	prompt := promptBuilder.BuildThinkingPrompt(description)

	// Create the request payload for Ollama API
	payload := map[string]interface{}{
		"model":  p.model,
		"prompt": prompt,
		"stream": false,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("thinking stage: failed to marshal request body: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", p.endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("thinking stage: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("thinking stage: API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("thinking stage: failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("thinking stage: API request failed with status %d: %s",
			resp.StatusCode, string(respBody))
	}

	// Parse the response
	var result OllamaResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("thinking stage: failed to parse API response: %w", err)
	}

	// Extract and parse JSON from the response
	extractedJSON, err := getJSONFromMessage(result.Response)
	if err != nil {
		return nil, fmt.Errorf("thinking stage: failed to extract JSON from response: %w", err)
	}

	// Parse the thinking output
	var thinking ThinkingOutput
	if err := json.Unmarshal([]byte(extractedJSON), &thinking); err != nil {
		return nil, fmt.Errorf("thinking stage: failed to parse thinking JSON: %w", err)
	}

	// Validate the response
	if len(thinking.Technologies) == 0 {
		return nil, errors.New("thinking stage: no technologies identified in response")
	}

	if thinking.ArchitecturePattern == "" {
		return nil, errors.New("thinking stage: no architecture pattern identified in response")
	}

	if len(thinking.KeyComponents) == 0 {
		return nil, errors.New("thinking stage: no key components identified in response")
	}

	return &thinking, nil
}

// ExecutePlanningStage executes the planning stage using Ollama's API.
// It takes the thinking output as context and returns a PlanningOutput.
func (p *OllamaProvider) ExecutePlanningStage(description string, thinking *ThinkingOutput) (*PlanningOutput, error) {
	if thinking == nil {
		return nil, errors.New("planning stage: thinking output is required")
	}

	promptBuilder := NewPromptBuilder()
	prompt := promptBuilder.BuildPlanningPrompt(description, thinking)

	// Create the request payload for Ollama API
	payload := map[string]interface{}{
		"model":  p.model,
		"prompt": prompt,
		"stream": false,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("planning stage: failed to marshal request body: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", p.endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("planning stage: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("planning stage: API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("planning stage: failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("planning stage: API request failed with status %d: %s",
			resp.StatusCode, string(respBody))
	}

	// Parse the response
	var result OllamaResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("planning stage: failed to parse API response: %w", err)
	}

	// Extract and parse JSON from the response
	extractedJSON, err := getJSONFromMessage(result.Response)
	if err != nil {
		return nil, fmt.Errorf("planning stage: failed to extract JSON from response: %w", err)
	}

	// Parse the planning output
	var planning PlanningOutput
	if err := json.Unmarshal([]byte(extractedJSON), &planning); err != nil {
		return nil, fmt.Errorf("planning stage: failed to parse planning JSON: %w", err)
	}

	// Validate the response
	if planning.ProjectName == "" {
		return nil, errors.New("planning stage: no project name in response")
	}

	if len(planning.DirectoryStructure) == 0 {
		return nil, errors.New("planning stage: empty directory structure in response")
	}

	return &planning, nil
}

// ExecuteGenerationStage executes the generation stage using Ollama's API.
// It takes both thinking and planning outputs as context and returns a BoilerplateResponse.
func (p *OllamaProvider) ExecuteGenerationStage(description string, thinking *ThinkingOutput, planning *PlanningOutput) (*BoilerplateResponse, error) {
	if thinking == nil {
		return nil, errors.New("generation stage: thinking output is required")
	}

	if planning == nil {
		return nil, errors.New("generation stage: planning output is required")
	}

	promptBuilder := NewPromptBuilder()
	prompt := promptBuilder.BuildGenerationPrompt(description, thinking, planning)

	// Create the request payload for Ollama API
	payload := map[string]interface{}{
		"model":  p.model,
		"prompt": prompt,
		"stream": false,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("generation stage: failed to marshal request body: %w", err)
	}

	url := fmt.Sprintf("%s/api/generate", p.endpoint)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("generation stage: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("generation stage: API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("generation stage: failed to read response body: %w", err)
	}

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("generation stage: API request failed with status %d: %s",
			resp.StatusCode, string(respBody))
	}

	// Parse the response
	var result OllamaResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("generation stage: failed to parse API response: %w", err)
	}

	// Extract and parse JSON from the response
	extractedJSON, err := getJSONFromMessage(result.Response)
	if err != nil {
		return nil, fmt.Errorf("generation stage: failed to extract JSON from response: %w", err)
	}

	// Parse the boilerplate response
	var boilerplate BoilerplateResponse
	if err := json.Unmarshal([]byte(extractedJSON), &boilerplate); err != nil {
		return nil, fmt.Errorf("generation stage: failed to parse boilerplate JSON: %w", err)
	}

	// Validate the response
	if len(boilerplate.ProjectStructure) == 0 {
		return nil, errors.New("generation stage: empty project structure in response")
	}

	if len(boilerplate.BoilerplateCode) == 0 {
		return nil, errors.New("generation stage: no boilerplate code in response")
	}

	return &boilerplate, nil
}

// GenerateBoilerplate implements Provider interface for Ollama
// GenerateBoilerplate generates project boilerplate using Ollama's API.
func (p *OllamaProvider) GenerateBoilerplate(description string) (*BoilerplateResponse, error) {
	url := fmt.Sprintf("%s/api/generate", p.endpoint)

	prompt := fmt.Sprintf(`Return as JSON. 
	Include: project_structure, boilerplate_code
	You are a software assistant that generates boilerplate code. 
	The sections (headers should be same) should include Project Structure (without comments), Boilerplate Code.
	The Project Structure should be represented as a nested json as a file system hierarchy and the file should have value as None
	Boilerplate Code file should directory mention from project path and it should be plain key-value pairs where key is file path and value is code text. 
	Validate the json and correct it.
	
	Generate boilerplate for a project described as: %s`, description)
	
	payload := map[string]interface{}{
		"model":  p.model,
		"prompt": prompt,
		"stream": false,
	}
	payloadBytes, _ := json.Marshal(payload)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API request failed with status: %s", resp.Status)
	}

	var result OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	extractedJSON, err := getJSONFromMessage(result.Response)
	if err != nil {
		return nil, fmt.Errorf("error extracting JSON from message: %w", err)
	}

	var boilerplate BoilerplateResponse
	if err := json.Unmarshal([]byte(extractedJSON), &boilerplate); err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	return &boilerplate, nil
}

// GenerationMetadata stores metadata about the generation process.
type GenerationMetadata struct {
	Thinking    *ThinkingOutput
	Planning    *PlanningOutput
	GeneratedAt time.Time
}

// StageOrchestrator coordinates the three-stage generation pipeline.
type StageOrchestrator struct {
	provider      Provider
	promptBuilder *PromptBuilder
}

// NewStageOrchestrator creates a new StageOrchestrator instance.
func NewStageOrchestrator(provider Provider) *StageOrchestrator {
	return &StageOrchestrator{
		provider:      provider,
		promptBuilder: NewPromptBuilder(),
	}
}

// GenerateWithStages orchestrates the three-stage generation process:
// thinking → planning → generation.
// It validates prerequisites between stages and returns both the final
// BoilerplateResponse and GenerationMetadata containing intermediate outputs.
func (so *StageOrchestrator) GenerateWithStages(description string) (*BoilerplateResponse, *GenerationMetadata, error) {
	// Stage 1: Thinking
	thinking, err := so.provider.ExecuteThinkingStage(description)
	if err != nil {
		return nil, nil, fmt.Errorf("thinking stage failed: %w", err)
	}

	// Validate thinking output before proceeding
	if thinking == nil {
		return nil, nil, errors.New("thinking stage returned nil output")
	}
	if len(thinking.Technologies) == 0 {
		return nil, nil, errors.New("thinking stage validation failed: no technologies identified")
	}
	if thinking.ArchitecturePattern == "" {
		return nil, nil, errors.New("thinking stage validation failed: no architecture pattern identified")
	}
	if len(thinking.KeyComponents) == 0 {
		return nil, nil, errors.New("thinking stage validation failed: no key components identified")
	}

	// Stage 2: Planning
	planning, err := so.provider.ExecutePlanningStage(description, thinking)
	if err != nil {
		return nil, nil, fmt.Errorf("planning stage failed: %w", err)
	}

	// Validate planning output before proceeding
	if planning == nil {
		return nil, nil, errors.New("planning stage returned nil output")
	}
	if planning.ProjectName == "" {
		return nil, nil, errors.New("planning stage validation failed: no project name")
	}
	if len(planning.DirectoryStructure) == 0 {
		return nil, nil, errors.New("planning stage validation failed: empty directory structure")
	}

	// Stage 3: Generation
	boilerplate, err := so.provider.ExecuteGenerationStage(description, thinking, planning)
	if err != nil {
		return nil, nil, fmt.Errorf("generation stage failed: %w", err)
	}

	// Validate generation output
	if boilerplate == nil {
		return nil, nil, errors.New("generation stage returned nil output")
	}
	if len(boilerplate.ProjectStructure) == 0 {
		return nil, nil, errors.New("generation stage validation failed: empty project structure")
	}
	if len(boilerplate.BoilerplateCode) == 0 {
		return nil, nil, errors.New("generation stage validation failed: no boilerplate code")
	}

	// Create metadata
	metadata := &GenerationMetadata{
		Thinking:    thinking,
		Planning:    planning,
		GeneratedAt: time.Now(),
	}

	return boilerplate, metadata, nil
}

// GetProvider creates and returns a new instance of the specified LLM provider.
// Supported providers are:
//   - "openai": Uses OpenAI's API (requires OPENAI_API_KEY)
//   - "claude": Uses Anthropic's Claude API (requires ANTHROPIC_API_KEY)
//   - "ollama": Uses a local Ollama instance (defaults to http://localhost:11434)
//
// Returns an error if the provider name is not supported or if initialization fails.
func GetProvider(name string) (Provider, error) {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "openai":
		return NewOpenAIProvider()
	case "claude":
		return NewClaudeProvider()
	case "ollama":
		return NewOllamaProvider()
	default:
		return nil, fmt.Errorf("unsupported provider: %s (supported: openai, claude, ollama)", name)
	}
}
