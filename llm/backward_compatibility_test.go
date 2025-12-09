package llm

import (
	"errors"
	"testing"
)

// TestGenerateBoilerplate_BackwardCompatibility tests that GenerateBoilerplate
// still works as expected and maintains the existing API contract.
// This test verifies that existing code using the old interface continues to work.
// Validates: Requirements 5.1
func TestGenerateBoilerplate_BackwardCompatibility(t *testing.T) {
	// Create a mock provider that implements the Provider interface
	mockProvider := &mockProviderForBackwardCompatibility{}

	// Test that GenerateBoilerplate returns a valid BoilerplateResponse
	response, err := mockProvider.GenerateBoilerplate("Create a simple web server")

	// Verify no error
	if err != nil {
		t.Fatalf("GenerateBoilerplate failed: %v", err)
	}

	// Verify response is not nil
	if response == nil {
		t.Fatal("GenerateBoilerplate returned nil response")
	}

	// Verify project structure is populated
	if len(response.ProjectStructure) == 0 {
		t.Error("ProjectStructure is empty")
	}

	// Verify boilerplate code is populated
	if len(response.BoilerplateCode) == 0 {
		t.Error("BoilerplateCode is empty")
	}
}

// TestGenerateBoilerplate_ReturnsOnlyBoilerplateResponse tests that
// GenerateBoilerplate returns only BoilerplateResponse without exposing
// intermediate outputs (thinking and planning).
// Validates: Requirements 5.1
func TestGenerateBoilerplate_ReturnsOnlyBoilerplateResponse(t *testing.T) {
	mockProvider := &mockProviderForBackwardCompatibility{}

	// Call GenerateBoilerplate
	response, err := mockProvider.GenerateBoilerplate("Create a REST API")

	if err != nil {
		t.Fatalf("GenerateBoilerplate failed: %v", err)
	}

	// Verify that the response is a BoilerplateResponse
	// and does not expose intermediate outputs
	if response == nil {
		t.Fatal("Response is nil")
	}

	// The response should only contain ProjectStructure and BoilerplateCode
	// No thinking or planning outputs should be exposed
	if response.ProjectStructure == nil {
		t.Error("ProjectStructure should not be nil")
	}

	if response.BoilerplateCode == nil {
		t.Error("BoilerplateCode should not be nil")
	}
}

// TestGenerateBoilerplate_UsesMultiStageInternally tests that
// GenerateBoilerplate internally uses the multi-stage generation process.
// Validates: Requirements 5.1
func TestGenerateBoilerplate_UsesMultiStageInternally(t *testing.T) {
	// Create a tracking provider that records which methods are called
	trackingProvider := &trackingProviderForBackwardCompatibility{
		thinkingCalled:   false,
		planningCalled:   false,
		generationCalled: false,
	}

	// Call GenerateBoilerplate
	_, err := trackingProvider.GenerateBoilerplate("Create a CLI tool")

	if err != nil {
		t.Fatalf("GenerateBoilerplate failed: %v", err)
	}

	// Verify that all three stages were called internally
	if !trackingProvider.thinkingCalled {
		t.Error("ExecuteThinkingStage was not called")
	}

	if !trackingProvider.planningCalled {
		t.Error("ExecutePlanningStage was not called")
	}

	if !trackingProvider.generationCalled {
		t.Error("ExecuteGenerationStage was not called")
	}
}

// TestGenerateBoilerplate_ErrorHandling tests that GenerateBoilerplate
// properly handles errors from the multi-stage generation process.
// Validates: Requirements 5.1
func TestGenerateBoilerplate_ErrorHandling(t *testing.T) {
	// Test with a provider that returns an error during thinking stage
	errorProvider := &errorProviderForBackwardCompatibility{
		failStage: "thinking",
	}

	_, err := errorProvider.GenerateBoilerplate("Create a web app")

	if err == nil {
		t.Error("Expected error from thinking stage, got nil")
	}

	// Test with a provider that returns an error during planning stage
	errorProvider.failStage = "planning"
	_, err = errorProvider.GenerateBoilerplate("Create a web app")

	if err == nil {
		t.Error("Expected error from planning stage, got nil")
	}

	// Test with a provider that returns an error during generation stage
	errorProvider.failStage = "generation"
	_, err = errorProvider.GenerateBoilerplate("Create a web app")

	if err == nil {
		t.Error("Expected error from generation stage, got nil")
	}
}

// TestGenerateBoilerplate_APIContract tests that the API contract
// is maintained for existing code.
// Validates: Requirements 5.1
func TestGenerateBoilerplate_APIContract(t *testing.T) {
	// This test verifies that the method signature hasn't changed
	// and that it can be called in the same way as before

	mockProvider := &mockProviderForBackwardCompatibility{}

	// Old usage pattern should still work
	var provider Provider = mockProvider
	response, err := provider.GenerateBoilerplate("test description")

	if err != nil {
		t.Fatalf("API contract broken: %v", err)
	}

	if response == nil {
		t.Fatal("API contract broken: nil response")
	}

	// Verify the response has the expected structure
	if response.ProjectStructure == nil {
		t.Error("API contract broken: ProjectStructure is nil")
	}

	if response.BoilerplateCode == nil {
		t.Error("API contract broken: BoilerplateCode is nil")
	}
}

// Mock providers for testing

// mockProviderForBackwardCompatibility is a mock provider that simulates
// the multi-stage generation process for backward compatibility testing.
type mockProviderForBackwardCompatibility struct{}

func (m *mockProviderForBackwardCompatibility) ExecuteThinkingStage(description string) (*ThinkingOutput, error) {
	return &ThinkingOutput{
		Technologies:        []string{"Go", "HTTP"},
		ArchitecturePattern: "REST API",
		KeyComponents:       []string{"Server", "Router"},
		Assumptions:         []string{"Using standard library"},
		Constraints:         []string{"Must be simple"},
	}, nil
}

func (m *mockProviderForBackwardCompatibility) ExecutePlanningStage(description string, thinking *ThinkingOutput) (*PlanningOutput, error) {
	return &PlanningOutput{
		ProjectName: "test-project",
		DirectoryStructure: map[string]interface{}{
			"test-project": map[string]interface{}{
				"main.go": nil,
			},
		},
		FileDescriptions: map[string]string{
			"test-project/main.go": "Main entry point",
		},
		Dependencies:       []Dependency{},
		ConfigurationFiles: []ConfigFile{},
	}, nil
}

func (m *mockProviderForBackwardCompatibility) ExecuteGenerationStage(description string, thinking *ThinkingOutput, planning *PlanningOutput) (*BoilerplateResponse, error) {
	return &BoilerplateResponse{
		ProjectStructure: planning.DirectoryStructure,
		BoilerplateCode: map[string]string{
			"test-project/main.go": "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}",
		},
	}, nil
}

func (m *mockProviderForBackwardCompatibility) GenerateBoilerplate(description string) (*BoilerplateResponse, error) {
	// Use StageOrchestrator to call multi-stage generation internally
	orchestrator := NewStageOrchestrator(m)
	boilerplate, _, err := orchestrator.GenerateWithStages(description)
	return boilerplate, err
}

// trackingProviderForBackwardCompatibility tracks which methods are called
type trackingProviderForBackwardCompatibility struct {
	thinkingCalled   bool
	planningCalled   bool
	generationCalled bool
}

func (t *trackingProviderForBackwardCompatibility) ExecuteThinkingStage(description string) (*ThinkingOutput, error) {
	t.thinkingCalled = true
	return &ThinkingOutput{
		Technologies:        []string{"Go"},
		ArchitecturePattern: "CLI",
		KeyComponents:       []string{"Parser"},
		Assumptions:         []string{"Simple CLI"},
		Constraints:         []string{"Minimal"},
	}, nil
}

func (t *trackingProviderForBackwardCompatibility) ExecutePlanningStage(description string, thinking *ThinkingOutput) (*PlanningOutput, error) {
	t.planningCalled = true
	return &PlanningOutput{
		ProjectName: "cli-tool",
		DirectoryStructure: map[string]interface{}{
			"cli-tool": map[string]interface{}{
				"main.go": nil,
			},
		},
		FileDescriptions:   map[string]string{"cli-tool/main.go": "CLI entry point"},
		Dependencies:       []Dependency{},
		ConfigurationFiles: []ConfigFile{},
	}, nil
}

func (t *trackingProviderForBackwardCompatibility) ExecuteGenerationStage(description string, thinking *ThinkingOutput, planning *PlanningOutput) (*BoilerplateResponse, error) {
	t.generationCalled = true
	return &BoilerplateResponse{
		ProjectStructure: planning.DirectoryStructure,
		BoilerplateCode: map[string]string{
			"cli-tool/main.go": "package main\n\nfunc main() {}",
		},
	}, nil
}

func (t *trackingProviderForBackwardCompatibility) GenerateBoilerplate(description string) (*BoilerplateResponse, error) {
	// Use StageOrchestrator to call multi-stage generation internally
	orchestrator := NewStageOrchestrator(t)
	boilerplate, _, err := orchestrator.GenerateWithStages(description)
	return boilerplate, err
}

// errorProviderForBackwardCompatibility simulates errors at different stages
type errorProviderForBackwardCompatibility struct {
	failStage string
}

func (e *errorProviderForBackwardCompatibility) ExecuteThinkingStage(description string) (*ThinkingOutput, error) {
	if e.failStage == "thinking" {
		return nil, errors.New("thinking stage error")
	}
	return &ThinkingOutput{
		Technologies:        []string{"Go"},
		ArchitecturePattern: "Web",
		KeyComponents:       []string{"Server"},
		Assumptions:         []string{"Test"},
		Constraints:         []string{"Test"},
	}, nil
}

func (e *errorProviderForBackwardCompatibility) ExecutePlanningStage(description string, thinking *ThinkingOutput) (*PlanningOutput, error) {
	if e.failStage == "planning" {
		return nil, errors.New("planning stage error")
	}
	return &PlanningOutput{
		ProjectName: "test",
		DirectoryStructure: map[string]interface{}{
			"test": map[string]interface{}{
				"main.go": nil,
			},
		},
		FileDescriptions:   map[string]string{"test/main.go": "Main"},
		Dependencies:       []Dependency{},
		ConfigurationFiles: []ConfigFile{},
	}, nil
}

func (e *errorProviderForBackwardCompatibility) ExecuteGenerationStage(description string, thinking *ThinkingOutput, planning *PlanningOutput) (*BoilerplateResponse, error) {
	if e.failStage == "generation" {
		return nil, errors.New("generation stage error")
	}
	return &BoilerplateResponse{
		ProjectStructure: planning.DirectoryStructure,
		BoilerplateCode:  map[string]string{"test/main.go": "package main"},
	}, nil
}

func (e *errorProviderForBackwardCompatibility) GenerateBoilerplate(description string) (*BoilerplateResponse, error) {
	// Use StageOrchestrator to call multi-stage generation internally
	orchestrator := NewStageOrchestrator(e)
	boilerplate, _, err := orchestrator.GenerateWithStages(description)
	return boilerplate, err
}
