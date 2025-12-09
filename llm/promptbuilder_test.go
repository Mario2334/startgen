package llm

import (
	"strings"
	"testing"
)

func TestBuildThinkingPrompt(t *testing.T) {
	pb := NewPromptBuilder()
	description := "A simple web server in Go"
	
	prompt := pb.BuildThinkingPrompt(description)
	
	// Verify prompt contains the description
	if !strings.Contains(prompt, description) {
		t.Error("Prompt should contain the project description")
	}
	
	// Verify prompt contains required analysis areas (Requirement 4.1)
	requiredKeywords := []string{"technologies", "architecture_pattern", "key_components", "assumptions", "constraints"}
	for _, keyword := range requiredKeywords {
		if !strings.Contains(prompt, keyword) {
			t.Errorf("Prompt should contain keyword: %s", keyword)
		}
	}
	
	// Verify prompt includes JSON format specification (Requirement 4.4)
	if !strings.Contains(prompt, "JSON format") {
		t.Error("Prompt should specify JSON format")
	}
	
	// Verify prompt includes field descriptions (Requirement 4.4)
	if !strings.Contains(prompt, "Field Descriptions") {
		t.Error("Prompt should include field descriptions")
	}
	
	// Verify prompt includes examples (Requirement 4.5)
	if !strings.Contains(prompt, "Example") {
		t.Error("Prompt should include examples")
	}
}

func TestBuildPlanningPrompt(t *testing.T) {
	pb := NewPromptBuilder()
	description := "A simple web server in Go"
	thinking := &ThinkingOutput{
		Technologies:        []string{"Go", "HTTP"},
		ArchitecturePattern: "REST API",
		KeyComponents:       []string{"Server", "Handlers"},
		Assumptions:         []string{"Using standard library"},
		Constraints:         []string{"Must be lightweight"},
	}
	
	prompt := pb.BuildPlanningPrompt(description, thinking)
	
	// Verify prompt contains the description
	if !strings.Contains(prompt, description) {
		t.Error("Prompt should contain the project description")
	}
	
	// Verify prompt includes thinking output as context (Requirement 4.2)
	if !strings.Contains(prompt, "REST API") {
		t.Error("Prompt should include thinking output context")
	}
	
	// Verify prompt requests structured plan (Requirement 4.2)
	if !strings.Contains(prompt, "implementation plan") {
		t.Error("Prompt should request implementation plan")
	}
	
	// Verify prompt includes JSON format specification (Requirement 4.4)
	if !strings.Contains(prompt, "JSON format") {
		t.Error("Prompt should specify JSON format")
	}
	
	// Verify prompt includes field descriptions (Requirement 4.4)
	if !strings.Contains(prompt, "Field Descriptions") {
		t.Error("Prompt should include field descriptions")
	}
	
	// Verify prompt includes examples (Requirement 4.5)
	if !strings.Contains(prompt, "Example") {
		t.Error("Prompt should include examples")
	}
}

func TestBuildGenerationPrompt(t *testing.T) {
	pb := NewPromptBuilder()
	description := "A simple web server in Go"
	thinking := &ThinkingOutput{
		Technologies:        []string{"Go", "HTTP"},
		ArchitecturePattern: "REST API",
		KeyComponents:       []string{"Server", "Handlers"},
		Assumptions:         []string{"Using standard library"},
		Constraints:         []string{"Must be lightweight"},
	}
	planning := &PlanningOutput{
		ProjectName: "go-web-server",
		DirectoryStructure: map[string]interface{}{
			"go-web-server": map[string]interface{}{
				"main.go": nil,
			},
		},
		FileDescriptions: map[string]string{
			"go-web-server/main.go": "Entry point",
		},
		Dependencies:       []Dependency{},
		ConfigurationFiles: []ConfigFile{},
	}
	
	prompt := pb.BuildGenerationPrompt(description, thinking, planning)
	
	// Verify prompt contains the description
	if !strings.Contains(prompt, description) {
		t.Error("Prompt should contain the project description")
	}
	
	// Verify prompt includes thinking output as context (Requirement 4.3)
	if !strings.Contains(prompt, "REST API") {
		t.Error("Prompt should include thinking output context")
	}
	
	// Verify prompt includes planning output as context (Requirement 4.3)
	if !strings.Contains(prompt, "go-web-server") {
		t.Error("Prompt should include planning output context")
	}
	
	// Verify prompt includes JSON format specification (Requirement 4.4)
	if !strings.Contains(prompt, "JSON format") {
		t.Error("Prompt should specify JSON format")
	}
	
	// Verify prompt includes field descriptions (Requirement 4.4)
	if !strings.Contains(prompt, "Field Descriptions") {
		t.Error("Prompt should include field descriptions")
	}
	
	// Verify prompt includes examples (Requirement 4.5)
	if !strings.Contains(prompt, "Example") {
		t.Error("Prompt should include examples")
	}
}
