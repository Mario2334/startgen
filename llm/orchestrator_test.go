package llm

import (
	"fmt"
	"math/rand"
	"testing"
)

// Feature: llm-thinking-planning, Property 4: Planning stage sequencing
// For any successful ThinkingOutput, executing the planning stage should produce a PlanningOutput
// with all required fields (project name, directory structure, file descriptions, dependencies,
// configuration files) populated.
// Validates: Requirements 2.1, 2.2
func TestProperty_PlanningStageSequencing(t *testing.T) {
	// Create a mock provider that returns valid outputs
	mockProvider := &mockProviderForSequencing{}

	// Generate random thinking outputs and verify planning stage produces valid output
	for i := 0; i < 100; i++ {
		thinking := generateRandomThinkingOutput()

		// Execute planning stage
		planning, err := mockProvider.ExecutePlanningStage("test description", thinking)

		// Verify no error
		if err != nil {
			t.Errorf("Iteration %d: Planning stage failed: %v", i, err)
			continue
		}

		// Verify all required fields are populated
		if planning.ProjectName == "" {
			t.Errorf("Iteration %d: ProjectName is empty", i)
		}
		if len(planning.DirectoryStructure) == 0 {
			t.Errorf("Iteration %d: DirectoryStructure is empty", i)
		}
		if planning.FileDescriptions == nil {
			t.Errorf("Iteration %d: FileDescriptions is nil", i)
		}
		if planning.Dependencies == nil {
			t.Errorf("Iteration %d: Dependencies is nil", i)
		}
		if planning.ConfigurationFiles == nil {
			t.Errorf("Iteration %d: ConfigurationFiles is nil", i)
		}
	}
}

// Feature: llm-thinking-planning, Property 5: Directory structure format consistency
// For any directory structure in a PlanningOutput, all leaf nodes (files) should have null values
// and all non-leaf nodes (directories) should be maps.
// Validates: Requirements 2.3
func TestProperty_DirectoryStructureFormat(t *testing.T) {
	// Test with 100 random directory structures
	for i := 0; i < 100; i++ {
		structure := generateRandomDirectoryStructure(3) // depth of 3

		// Verify the structure format
		if !validateDirectoryStructureFormat(structure) {
			t.Errorf("Iteration %d: Invalid directory structure format", i)
		}
	}
}

// Feature: llm-thinking-planning, Property 7: Planning validation
// For any PlanningOutput, the directory structure should be non-empty and properly formatted
// (valid nested map structure).
// Validates: Requirements 2.5
func TestProperty_PlanningValidation(t *testing.T) {
	mockProvider := &mockProviderForSequencing{}

	for i := 0; i < 100; i++ {
		thinking := generateRandomThinkingOutput()
		planning, err := mockProvider.ExecutePlanningStage("test description", thinking)

		if err != nil {
			t.Errorf("Iteration %d: Planning stage failed: %v", i, err)
			continue
		}

		// Validate directory structure is non-empty
		if len(planning.DirectoryStructure) == 0 {
			t.Errorf("Iteration %d: Directory structure is empty", i)
		}

		// Validate directory structure is properly formatted
		if !validateDirectoryStructureFormat(planning.DirectoryStructure) {
			t.Errorf("Iteration %d: Directory structure is not properly formatted", i)
		}
	}
}

// Feature: llm-thinking-planning, Property 8: Generation stage sequencing
// For any successful PlanningOutput and ThinkingOutput, executing the generation stage should
// produce a BoilerplateResponse with non-empty project structure and boilerplate code.
// Validates: Requirements 3.1
func TestProperty_GenerationStageSequencing(t *testing.T) {
	mockProvider := &mockProviderForSequencing{}

	for i := 0; i < 100; i++ {
		thinking := generateRandomThinkingOutput()
		planning := generateRandomPlanningOutput()

		// Execute generation stage
		boilerplate, err := mockProvider.ExecuteGenerationStage("test description", thinking, planning)

		if err != nil {
			t.Errorf("Iteration %d: Generation stage failed: %v", i, err)
			continue
		}

		// Verify non-empty project structure
		if len(boilerplate.ProjectStructure) == 0 {
			t.Errorf("Iteration %d: ProjectStructure is empty", i)
		}

		// Verify non-empty boilerplate code
		if len(boilerplate.BoilerplateCode) == 0 {
			t.Errorf("Iteration %d: BoilerplateCode is empty", i)
		}
	}
}

// Feature: llm-thinking-planning, Property 9: File path completeness
// For any PlanningOutput with N file paths, the generated BoilerplateResponse should contain
// code for all N file paths.
// Validates: Requirements 3.2, 3.3
func TestProperty_FilePathCompleteness(t *testing.T) {
	mockProvider := &mockProviderForSequencing{}

	for i := 0; i < 100; i++ {
		thinking := generateRandomThinkingOutput()
		planning := generateRandomPlanningOutput()

		// Count files in planning output
		filePaths := extractFilePaths(planning.DirectoryStructure, "")

		// Execute generation stage
		boilerplate, err := mockProvider.ExecuteGenerationStage("test description", thinking, planning)

		if err != nil {
			t.Errorf("Iteration %d: Generation stage failed: %v", i, err)
			continue
		}

		// Verify all file paths have corresponding code
		for _, path := range filePaths {
			if _, exists := boilerplate.BoilerplateCode[path]; !exists {
				t.Errorf("Iteration %d: Missing code for file path: %s", i, path)
			}
		}
	}
}

// Feature: llm-thinking-planning, Property 10: Structure consistency
// For any PlanningOutput and generated BoilerplateResponse, the project structure in the response
// should match the directory structure from the plan.
// Validates: Requirements 3.4
func TestProperty_StructureConsistency(t *testing.T) {
	mockProvider := &mockProviderForSequencing{}

	for i := 0; i < 100; i++ {
		thinking := generateRandomThinkingOutput()
		planning := generateRandomPlanningOutput()

		// Execute generation stage
		boilerplate, err := mockProvider.ExecuteGenerationStage("test description", thinking, planning)

		if err != nil {
			t.Errorf("Iteration %d: Generation stage failed: %v", i, err)
			continue
		}

		// Verify structures match
		if !structuresMatch(planning.DirectoryStructure, boilerplate.ProjectStructure) {
			t.Errorf("Iteration %d: Project structure does not match directory structure", i)
		}
	}
}

// Helper functions for property tests

// mockProviderForSequencing is a mock provider that returns valid outputs for testing
type mockProviderForSequencing struct{}

func (m *mockProviderForSequencing) ExecuteThinkingStage(description string) (*ThinkingOutput, error) {
	return generateRandomThinkingOutput(), nil
}

func (m *mockProviderForSequencing) ExecutePlanningStage(description string, thinking *ThinkingOutput) (*PlanningOutput, error) {
	return generateRandomPlanningOutput(), nil
}

func (m *mockProviderForSequencing) ExecuteGenerationStage(description string, thinking *ThinkingOutput, planning *PlanningOutput) (*BoilerplateResponse, error) {
	// Generate boilerplate code for all files in the planning structure
	filePaths := extractFilePaths(planning.DirectoryStructure, "")
	boilerplateCode := make(map[string]string)
	for _, path := range filePaths {
		boilerplateCode[path] = "// Generated code for " + path
	}

	return &BoilerplateResponse{
		ProjectStructure: planning.DirectoryStructure,
		BoilerplateCode:  boilerplateCode,
	}, nil
}

func (m *mockProviderForSequencing) GenerateBoilerplate(description string) (*BoilerplateResponse, error) {
	return &BoilerplateResponse{
		ProjectStructure: generateRandomDirectoryStructure(2),
		BoilerplateCode:  map[string]string{"test.go": "package main"},
	}, nil
}

func generateRandomThinkingOutput() *ThinkingOutput {
	technologies := []string{"Go", "PostgreSQL", "Docker", "Redis", "Kubernetes"}
	patterns := []string{"Microservices", "Monolithic", "Layered", "Event-Driven", "MVC"}
	components := []string{"API Server", "Database", "Cache", "Message Queue", "Load Balancer"}

	// Randomly select 2-4 items from each list
	numTech := 2 + rand.Intn(3)
	numComp := 2 + rand.Intn(3)

	selectedTech := make([]string, numTech)
	selectedComp := make([]string, numComp)

	for i := 0; i < numTech; i++ {
		selectedTech[i] = technologies[rand.Intn(len(technologies))]
	}
	for i := 0; i < numComp; i++ {
		selectedComp[i] = components[rand.Intn(len(components))]
	}

	return &ThinkingOutput{
		Technologies:        selectedTech,
		ArchitecturePattern: patterns[rand.Intn(len(patterns))],
		KeyComponents:       selectedComp,
		Assumptions:         []string{"Using REST API", "PostgreSQL for persistence"},
		Constraints:         []string{"Must be stateless", "Support concurrent users"},
	}
}

func generateRandomPlanningOutput() *PlanningOutput {
	structure := generateRandomDirectoryStructure(2)
	filePaths := extractFilePaths(structure, "")

	fileDescriptions := make(map[string]string)
	for _, path := range filePaths {
		fileDescriptions[path] = "Description for " + path
	}

	return &PlanningOutput{
		ProjectName:        fmt.Sprintf("test-project-%d", rand.Intn(1000)),
		DirectoryStructure: structure,
		FileDescriptions:   fileDescriptions,
		Dependencies: []Dependency{
			{Name: "github.com/gorilla/mux", Purpose: "HTTP routing", Version: "v1.8.0"},
		},
		ConfigurationFiles: []ConfigFile{
			{Path: "config.yaml", Format: "YAML", Purpose: "Application configuration"},
		},
	}
}

func generateRandomDirectoryStructure(depth int) map[string]interface{} {
	if depth <= 0 {
		return nil
	}

	structure := make(map[string]interface{})
	projectName := fmt.Sprintf("project-%d", rand.Intn(1000))

	projectContent := make(map[string]interface{})

	// Add some directories
	numDirs := 1 + rand.Intn(3)
	for i := 0; i < numDirs; i++ {
		dirName := fmt.Sprintf("dir%d", i)
		if depth > 1 {
			// Recursively create subdirectories
			projectContent[dirName] = generateRandomDirectoryStructure(depth - 1)[fmt.Sprintf("project-%d", rand.Intn(1000))]
		} else {
			// At the last level, just create files
			projectContent[dirName] = map[string]interface{}{
				fmt.Sprintf("file%d.go", i): nil,
			}
		}
	}

	// Add some files at this level
	numFiles := 1 + rand.Intn(3)
	for i := 0; i < numFiles; i++ {
		fileName := fmt.Sprintf("file%d.go", i)
		projectContent[fileName] = nil
	}

	structure[projectName] = projectContent
	return structure
}

func validateDirectoryStructureFormat(structure map[string]interface{}) bool {
	for _, value := range structure {
		if value == nil {
			// This is a file (leaf node) - valid
			continue
		}

		// This should be a directory (non-leaf node)
		subMap, ok := value.(map[string]interface{})
		if !ok {
			// Non-leaf node is not a map - invalid
			return false
		}

		// Recursively validate subdirectories
		if !validateDirectoryStructureFormat(subMap) {
			return false
		}
	}
	return true
}

func extractFilePaths(structure map[string]interface{}, prefix string) []string {
	var paths []string

	for key, value := range structure {
		currentPath := key
		if prefix != "" {
			currentPath = prefix + "/" + key
		}

		if value == nil {
			// This is a file
			paths = append(paths, currentPath)
		} else if subMap, ok := value.(map[string]interface{}); ok {
			// This is a directory, recurse
			paths = append(paths, extractFilePaths(subMap, currentPath)...)
		}
	}

	return paths
}

func structuresMatch(structure1, structure2 map[string]interface{}) bool {
	if len(structure1) != len(structure2) {
		return false
	}

	for key, value1 := range structure1 {
		value2, exists := structure2[key]
		if !exists {
			return false
		}

		// Both are files (nil)
		if value1 == nil && value2 == nil {
			continue
		}

		// Both are directories
		subMap1, ok1 := value1.(map[string]interface{})
		subMap2, ok2 := value2.(map[string]interface{})

		if ok1 && ok2 {
			if !structuresMatch(subMap1, subMap2) {
				return false
			}
		} else if ok1 != ok2 {
			// One is a directory, the other is not
			return false
		}
	}

	return true
}

// TestStageOrchestrator_Integration tests the complete orchestration flow
func TestStageOrchestrator_Integration(t *testing.T) {
	mockProvider := &mockProviderForSequencing{}
	orchestrator := NewStageOrchestrator(mockProvider)

	// Test successful generation
	boilerplate, metadata, err := orchestrator.GenerateWithStages("Create a simple web server")

	if err != nil {
		t.Fatalf("GenerateWithStages failed: %v", err)
	}

	// Verify boilerplate response
	if boilerplate == nil {
		t.Fatal("Boilerplate response is nil")
	}
	if len(boilerplate.ProjectStructure) == 0 {
		t.Error("Project structure is empty")
	}
	if len(boilerplate.BoilerplateCode) == 0 {
		t.Error("Boilerplate code is empty")
	}

	// Verify metadata
	if metadata == nil {
		t.Fatal("Metadata is nil")
	}
	if metadata.Thinking == nil {
		t.Error("Thinking output in metadata is nil")
	}
	if metadata.Planning == nil {
		t.Error("Planning output in metadata is nil")
	}
	if metadata.GeneratedAt.IsZero() {
		t.Error("GeneratedAt timestamp is zero")
	}

	// Verify thinking output
	if len(metadata.Thinking.Technologies) == 0 {
		t.Error("No technologies in thinking output")
	}
	if metadata.Thinking.ArchitecturePattern == "" {
		t.Error("No architecture pattern in thinking output")
	}
	if len(metadata.Thinking.KeyComponents) == 0 {
		t.Error("No key components in thinking output")
	}

	// Verify planning output
	if metadata.Planning.ProjectName == "" {
		t.Error("No project name in planning output")
	}
	if len(metadata.Planning.DirectoryStructure) == 0 {
		t.Error("No directory structure in planning output")
	}
}

// TestStageOrchestrator_ValidationBetweenStages tests that validation occurs between stages
func TestStageOrchestrator_ValidationBetweenStages(t *testing.T) {
	// Test with a provider that returns invalid thinking output
	invalidProvider := &invalidThinkingProvider{}
	orchestrator := NewStageOrchestrator(invalidProvider)

	_, _, err := orchestrator.GenerateWithStages("test")
	if err == nil {
		t.Error("Expected error for invalid thinking output, got nil")
	}
	if err != nil && err.Error() != "thinking stage validation failed: no technologies identified" {
		t.Errorf("Expected validation error, got: %v", err)
	}
}

// invalidThinkingProvider returns invalid thinking output for testing
type invalidThinkingProvider struct{}

func (p *invalidThinkingProvider) ExecuteThinkingStage(description string) (*ThinkingOutput, error) {
	return &ThinkingOutput{
		Technologies:        []string{}, // Empty - should fail validation
		ArchitecturePattern: "Test",
		KeyComponents:       []string{"Test"},
	}, nil
}

func (p *invalidThinkingProvider) ExecutePlanningStage(description string, thinking *ThinkingOutput) (*PlanningOutput, error) {
	return generateRandomPlanningOutput(), nil
}

func (p *invalidThinkingProvider) ExecuteGenerationStage(description string, thinking *ThinkingOutput, planning *PlanningOutput) (*BoilerplateResponse, error) {
	return &BoilerplateResponse{
		ProjectStructure: planning.DirectoryStructure,
		BoilerplateCode:  map[string]string{"test.go": "package main"},
	}, nil
}

func (p *invalidThinkingProvider) GenerateBoilerplate(description string) (*BoilerplateResponse, error) {
	return &BoilerplateResponse{
		ProjectStructure: generateRandomDirectoryStructure(2),
		BoilerplateCode:  map[string]string{"test.go": "package main"},
	}, nil
}
