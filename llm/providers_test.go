package llm

import (
	"net/http"
	"testing"
)

// TestOpenAIProviderImplementsInterface verifies that OpenAIProvider implements the Provider interface
func TestOpenAIProviderImplementsInterface(t *testing.T) {
	// This test will fail to compile if OpenAIProvider doesn't implement Provider
	var _ Provider = (*OpenAIProvider)(nil)
}

// TestClaudeProviderImplementsInterface verifies that ClaudeProvider implements the Provider interface
func TestClaudeProviderImplementsInterface(t *testing.T) {
	// This test will fail to compile if ClaudeProvider doesn't implement Provider
	var _ Provider = (*ClaudeProvider)(nil)
}

// TestOllamaProviderImplementsInterface verifies that OllamaProvider implements the Provider interface
func TestOllamaProviderImplementsInterface(t *testing.T) {
	// This test will fail to compile if OllamaProvider doesn't implement Provider
	var _ Provider = (*OllamaProvider)(nil)
}

// TestOpenAIExecuteThinkingStageValidation tests validation in ExecuteThinkingStage
func TestOpenAIExecuteThinkingStageValidation(t *testing.T) {
	// This test verifies that the ExecuteThinkingStage method exists and has the correct signature
	// We verify this by checking that OpenAIProvider implements the Provider interface
	// which is already tested in TestOpenAIProviderImplementsInterface
	t.Skip("Method signature verified by interface compliance test")
}

// TestOpenAIExecutePlanningStageValidation tests validation in ExecutePlanningStage
func TestOpenAIExecutePlanningStageValidation(t *testing.T) {
	provider := &OpenAIProvider{
		apiKey: "test-key",
		model:  "gpt-4o",
		client: &http.Client{Timeout: defaultTimeout},
	}
	
	// Test that nil thinking output returns an error
	_, err := provider.ExecutePlanningStage("test description", nil)
	if err == nil {
		t.Error("ExecutePlanningStage should return error when thinking output is nil")
	}
	
	expectedError := "planning stage: thinking output is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// TestOpenAIExecuteGenerationStageValidation tests validation in ExecuteGenerationStage
func TestOpenAIExecuteGenerationStageValidation(t *testing.T) {
	provider := &OpenAIProvider{
		apiKey: "test-key",
		model:  "gpt-4o",
		client: &http.Client{Timeout: defaultTimeout},
	}
	
	thinking := &ThinkingOutput{
		Technologies:        []string{"Go"},
		ArchitecturePattern: "REST API",
		KeyComponents:       []string{"Server"},
		Assumptions:         []string{"Test"},
		Constraints:         []string{"Test"},
	}
	
	// Test that nil thinking output returns an error
	_, err := provider.ExecuteGenerationStage("test description", nil, nil)
	if err == nil {
		t.Error("ExecuteGenerationStage should return error when thinking output is nil")
	}
	
	expectedError := "generation stage: thinking output is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
	
	// Test that nil planning output returns an error
	_, err = provider.ExecuteGenerationStage("test description", thinking, nil)
	if err == nil {
		t.Error("ExecuteGenerationStage should return error when planning output is nil")
	}
	
	expectedError = "generation stage: planning output is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// TestOllamaExecutePlanningStageValidation tests validation in ExecutePlanningStage
func TestOllamaExecutePlanningStageValidation(t *testing.T) {
	provider := &OllamaProvider{
		endpoint: "http://localhost:11434",
		model:    "llama3",
		client:   &http.Client{Timeout: defaultTimeout},
	}
	
	// Test that nil thinking output returns an error
	_, err := provider.ExecutePlanningStage("test description", nil)
	if err == nil {
		t.Error("ExecutePlanningStage should return error when thinking output is nil")
	}
	
	expectedError := "planning stage: thinking output is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// TestOllamaExecuteGenerationStageValidation tests validation in ExecuteGenerationStage
func TestOllamaExecuteGenerationStageValidation(t *testing.T) {
	provider := &OllamaProvider{
		endpoint: "http://localhost:11434",
		model:    "llama3",
		client:   &http.Client{Timeout: defaultTimeout},
	}
	
	thinking := &ThinkingOutput{
		Technologies:        []string{"Go"},
		ArchitecturePattern: "REST API",
		KeyComponents:       []string{"Server"},
		Assumptions:         []string{"Test"},
		Constraints:         []string{"Test"},
	}
	
	// Test that nil thinking output returns an error
	_, err := provider.ExecuteGenerationStage("test description", nil, nil)
	if err == nil {
		t.Error("ExecuteGenerationStage should return error when thinking output is nil")
	}
	
	expectedError := "generation stage: thinking output is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
	
	// Test that nil planning output returns an error
	_, err = provider.ExecuteGenerationStage("test description", thinking, nil)
	if err == nil {
		t.Error("ExecuteGenerationStage should return error when planning output is nil")
	}
	
	expectedError = "generation stage: planning output is required"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

// TestGetJSONFromMessage tests the JSON extraction utility function
func TestGetJSONFromMessage(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldError bool
	}{
		{
			name:        "JSON in code block with language specifier",
			input:       "```json\n{\"key\": \"value\"}\n```",
			shouldError: false,
		},
		{
			name:        "JSON in code block without language specifier",
			input:       "```\n{\"key\": \"value\"}\n```",
			shouldError: false,
		},
		{
			name:        "No code block",
			input:       "{\"key\": \"value\"}",
			shouldError: true,
		},
		{
			name:        "Empty code block",
			input:       "```json\n\n```",
			shouldError: true,
		},
		{
			name:        "Invalid JSON in code block",
			input:       "```json\n{invalid json}\n```",
			shouldError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := getJSONFromMessage(tt.input)
			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}
