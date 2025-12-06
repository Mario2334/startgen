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
	defaultTimeout = 60 * time.Second
)

// Provider defines the interface that all LLM providers must implement.
type Provider interface {
	// GenerateBoilerplate generates project boilerplate based on the given description.
	// It returns a BoilerplateResponse containing the project structure and boilerplate code,
	// or an error if the generation fails.
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

// BuildThinkingPrompt creates a prompt for the thinking stage that analyzes pro
	// First try to find a JSON code block with language specifier
	jsonPattern := regexp.MustCompile("(?s)```json\\s*(.*?)```")
	matches := jsonPattern.FindStringSubmatch(data)

	// If no match, try to find any code block
	if len(matches) < 2 {
		codeBlockPattern := regexp.MustCompile("(?s)```(.*?)```")
		matches = codeBlockPattern.FindStringSubmatch(data)
	}

	if len(matches) < 2 {
		return "", errors.New("no code block found in the response")
	}

	extracted := strings.TrimSpace(matches[1])
	if extracted == "" {
		return "", errors.New("empty content in code block")
	}

	// Validate JSON syntax
	var js map[string]interface{}
	if err := json.Unmarshal([]byte(extracted), &js); err != nil {
		return "", fmt.Errorf("invalid JSON in response: %w", err)
	}

	return extracted, nil
}

// OpenAIMessage represents a message in the OpenAI chat format.
type OpenAIMessage struct {
	Role    string `json:"role"`    // "system", "user", or "assistant"
	Content string `json:"content"` // The actual message content
}

// OpenAICompletionRequest represents a request to the OpenAI chat completion API.
type OpenAICompletionRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
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
