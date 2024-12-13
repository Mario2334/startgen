package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/briandowns/spinner"
)

var (
	openAIAPIKey string
)

// Message Structure for the Message field inside Choices
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Choice Structure for individual Choice
type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason,omitempty"`
}

// AssistantResponse Main structure to represent the entire JSON
type AssistantResponse struct {
	ID                string   `json:"id"`
	Object            string   `json:"object"`
	Created           int64    `json:"created"`
	Model             string   `json:"model"`
	Choices           []Choice `json:"choices"`
	SystemFingerprint string   `json:"system_fingerprint"`
}

type BoilerplateResponse struct {
	ProjectStructure map[string]interface{} `json:"project_structure"`
	BoilerplateCode  map[string]string      `json:"boilerplate_code"`
}

func init() {
	openAIAPIKey = os.Getenv("OPENAI_API_KEY")

	if openAIAPIKey == "" {
		fmt.Println("❌ Error: The environment variable 'OPENAI_API_KEY' is not set. Please set it to use the OpenAI API.")
		os.Exit(1)
	}
}

func getJsonFromMessage(data string) (*string, error) {
	jsonPattern := regexp.MustCompile("(?s)```json\\n(.*?)\\n```")

	// Extract JSON from the input text
	matches := jsonPattern.FindStringSubmatch(data)
	if len(matches) < 2 {
		fmt.Println("No JSON found in the input text.")
		return nil, errors.New("No JSON found in the input text.")
	}
	extractedJSON := matches[1]
	return &extractedJSON, nil
}

func generateBoilerplate(description string) (*BoilerplateResponse, error) {
	url := "https://api.openai.com/v1/chat/completions" // Replace with actual OpenAI endpoint
	message_data := []map[string]interface{}{
		{
			"role": "system",
			"content": `Return as JSON. 
        Include:  project_structure, boilerplate_code
        You are a software assistant that generates boilerplate code. 
        The sections (headers should be same) should include Project Structure (without comments), Boilerplate Code.
        The Project Structure should be represented as a nested json as a file system hierarchy and the file should have value as None
        Boilerplate Code file should directory mention from project path and it should be plain key-value pairs where key is file path and value is code text. 
        Validate the json and correct it`,
		},
		{
			"role":    "user",
			"content": fmt.Sprintf("Generate boilerplate for a project described as: %s", description),
		},
	}
	payload := map[string]interface{}{
		"model":    "gpt-4o",
		"messages": message_data,
	}
	payloadBytes, _ := json.Marshal(payload)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", openAIAPIKey))
	req.Header.Set("Content-Type", "application/json")
	req.Body = ioutil.NopCloser(bytes.NewBuffer(payloadBytes))

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bytedata, _ := io.ReadAll(resp.Body)
		reqBodyString := string(bytedata)

		return nil, fmt.Errorf("API request failed with status: %s", reqBodyString)
	}

	var result AssistantResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}
	message := result.Choices[0].Message.Content
	extractedJSON, err := getJsonFromMessage(message)

	if err != nil {
		return nil, err
	}

	var boilerplateResponse BoilerplateResponse
	json.Unmarshal([]byte(*extractedJSON), &boilerplateResponse)

	return &boilerplateResponse, nil
}

func createFilesFromStructure(structure map[string]interface{}, parentDir string) error {
	for name, value := range structure {
		currentPath := filepath.Join(parentDir, name)
		if subStructure, ok := value.(map[string]interface{}); ok {
			err := os.MkdirAll(currentPath, os.ModePerm)
			if err != nil {
				return err
			}
			err = createFilesFromStructure(subStructure, currentPath)
			if err != nil {
				return err
			}
		} else {
			err := ioutil.WriteFile(currentPath, []byte{}, 0644)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	// Parse CLI arguments
	outputDir := flag.String("output-dir", ".", "Directory to save the generated boilerplate.")
	print(outputDir)
	flag.Parse()

	// Positional argument for the project description
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("❌ Error: Please provide a project description as a positional argument.")
		flag.Usage()
		os.Exit(1)
	}
	description := args[0]

	// Display progress spinner
	spin := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	spin.Start()
	spin.Suffix = " Generating boilerplate..."

	boilerplate, err := generateBoilerplate(description)
	print(boilerplate)
	spin.Stop()
	if err != nil {
		fmt.Printf("❌ Error generating boilerplate: %v\n", err)
		os.Exit(1)
	}

	//Create files from structure
	err = createFilesFromStructure(boilerplate.ProjectStructure, *outputDir)
	if err != nil {
		fmt.Printf("❌ Error creating files: %v\n", err)
		os.Exit(1)
	}

	//Write boilerplate code to files
	for filePath, content := range boilerplate.BoilerplateCode {
		fullPath := filepath.Join(*outputDir, filePath)
		err := ioutil.WriteFile(fullPath, []byte(content), 0644)
		if err != nil {
			fmt.Printf("❌ Error writing to file %s: %v\n", fullPath, err)
			os.Exit(1)
		}
	}

	fmt.Println("🚀 Project files have been generated successfully!")
}
