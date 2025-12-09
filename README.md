
# 🚀 StartGen

**StartGen** is an AI-powered project scaffolding tool that brings your project ideas to life. Describe your project in plain English, and StartGen generates a complete, well-structured codebase ready for development.

---

## Features

- **Plain English Prompts:** Describe your project in natural language
- **Multi-Stage Generation:** Three-stage process (Thinking, Planning, Generation) ensures thoughtful project design
- **Multiple LLM Providers:** Support for OpenAI, Claude, and Ollama
- **Clean Project Structure:** Well-organized directories and files
- **Generation Metadata:** View technologies, architecture patterns, and components identified for your project
- **Flexible Output:** Choose where to generate your project

---

## Installation

### Prerequisites

Set up your LLM provider credentials:

**For OpenAI:**
```bash
export OPENAI_API_KEY=your_openai_api_key_here
```

**For Claude:**
```bash
export ANTHROPIC_API_KEY=your_anthropic_api_key_here
```

**For Ollama (optional):**
```bash
export OLLAMA_ENDPOINT=http://localhost:11434  # defaults to this if not set
export OLLAMA_MODEL=llama3                      # defaults to llama3 if not set
```

### Install via Homebrew

```bash
brew tap Mario2334/startgen
brew install startgen
```

### Install via Binary

1. Download the `startgen` binary for your system from the [releases page](https://github.com/Mario2334/startgen/releases)

2. Make it executable:
   ```bash
   chmod +x ./startgen
   ```

3. Optionally, move it to your PATH:
   ```bash
   sudo mv ./startgen /usr/local/bin/
   ```

### Build from Source

**Prerequisites:**
- Go 1.21 or higher
- Git

**Steps:**

1. Clone the repository:
   ```bash
   git clone https://github.com/Mario2334/startgen.git
   cd startgen
   ```

2. Build the project:
   ```bash
   go build -o startgen .
   ```

3. Run StartGen:
   ```bash
   ./startgen "Your project description here"
   ```

   Or run directly without building:
   ```bash
   go run . "Your project description here"
   ```

---

## Usage

### Basic Usage

```bash
startgen "A React app with a Flask backend and PostgreSQL database"
```

### Specify Output Directory

```bash
startgen -o ~/projects "A Node.js CLI tool that converts CSV to JSON"
```

### Use a Different Provider

```bash
startgen -p claude "A Python web scraper with BeautifulSoup"
```

### Save Generation Metadata

```bash
startgen --save-metadata "A REST API with authentication"
```

This saves the thinking and planning outputs to `.startgen-metadata.json` in your project directory.

### View Available Providers

```bash
startgen -l
```

### Get Help

```bash
startgen --help
```

---

## Command-Line Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--help` | `-h` | Show help message | - |
| `--provider` | `-p` | LLM provider to use (openai, claude, ollama) | openai |
| `--output-dir` | `-o` | Directory to save the generated project | . (current directory) |
| `--providers` | `-l` | Show information about available providers | - |
| `--save-metadata` | - | Save thinking and planning outputs to `.startgen-metadata.json` | false |

---

## How Multi-Stage Generation Works

StartGen uses a three-stage approach to ensure high-quality, well-structured projects:

### Stage 1: Thinking

The AI analyzes your project description to identify:
- Technologies and frameworks needed
- Appropriate architecture patterns
- Key components and their relationships
- Assumptions about unclear requirements
- Technical constraints

**Example output:** For a "React app with Flask backend", the thinking stage identifies React, Flask, PostgreSQL, REST API patterns, and component structure.

### Stage 2: Planning

Based on the thinking stage analysis, the AI creates:
- Complete directory structure
- Detailed file descriptions
- Required dependencies with versions
- Configuration file specifications

**Example output:** A detailed plan showing `frontend/`, `backend/`, and `database/` directories with specific files like `App.js`, `routes.py`, and `setup.sql`.

### Stage 3: Generation

Finally, the AI generates:
- All source code files with working boilerplate
- Configuration files (package.json, requirements.txt, etc.)
- Documentation files
- Setup scripts

**Example output:** Fully functional code files ready for development.

Each stage builds on the previous one, ensuring your project is thoughtfully designed before any code is written. You can view the thinking and planning outputs by using the `--save-metadata` flag.

---

## Example Workflow

1. **Describe your project:**
   ```bash
   startgen "A Python-based data pipeline to process weather data"
   ```

2. **StartGen processes your request:**
   - **Thinking:** Analyzes requirements and identifies Python, Apache Airflow, pandas, and data processing patterns
   - **Planning:** Creates directory structure with `dags/`, `scripts/`, and `config/` folders
   - **Generation:** Generates all code files

3. **Review the output:**
   ```
   🎉 Success! Your Project Has Been Generated 🚀
   --------------------------------------------
   🔸 Provider Used: openai
   🔸 Project Name: weather-data-pipeline
   🔸 Project Path: ./weather-data-pipeline
   
   📋 Generation Metadata:
     Technologies:
       • Python
       • Apache Airflow
       • pandas
     Architecture Pattern: ETL Pipeline
     Key Components:
       • Data ingestion module
       • Transformation scripts
       • Airflow DAGs
   --------------------------------------------
   ```

4. **Start coding:**
   ```bash
   cd weather-data-pipeline
   # Your project is ready for development!
   ```

---

## Troubleshooting

### Provider Authentication Errors

**Error:** "Error initializing provider"

**Solution:** Ensure you have set the correct environment variable for your chosen provider:
- OpenAI: `export OPENAI_API_KEY=your_key`
- Claude: `export ANTHROPIC_API_KEY=your_key`
- Ollama: Ensure Ollama is running locally (default: `http://localhost:11434`)

Use `startgen -l` to see provider-specific requirements.

### No Output or Empty Project

**Issue:** StartGen completes but generates no files or an empty project.

**Solution:**
- Make your prompt more specific and descriptive
- Include key technologies you want to use
- Specify the type of project (web app, CLI tool, API, etc.)

**Example of a good prompt:**
```bash
startgen "A REST API using Flask with PostgreSQL database, including user authentication and CRUD operations for blog posts"
```

### Unexpected Project Structure

**Issue:** The generated project doesn't match your expectations.

**Solution:**
- Refine your prompt with more details about the desired structure
- Specify frameworks, libraries, or patterns you want to use
- Try using `--save-metadata` to review what the AI understood from your prompt

### Permission Denied Errors

**Issue:** Cannot create files or directories.

**Solution:**
- Ensure you have write permissions in the output directory
- Try specifying a different output directory with `-o ~/projects`
- Check that the output directory exists and is accessible

---
## Support & Feedback

Got questions or suggestions? Open an issue on the [GitHub repository](https://github.com/Mario2334/startgen).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is open source and available under the MIT License.

---

**StartGen** - Transform ideas into code, effortlessly.
