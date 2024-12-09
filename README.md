# StartGen

**StartGen** is your fun, AI-powered friend that quickly spins up a boilerplate project structure from your description. No more head-scratching over file hierarchies—just tell StartGen what you want to build, grab a coffee, and watch the magic happen!

## What Does StartGen Do?

StartGen takes a natural language prompt—such as:  
*"I want a Flask-based API that manages a list of heroes and villains!"*  
—and transforms it into a neat directory structure and some starter files. You’ll be coding productively in no time!

## Key Features

- **Plain English Prompts:** Simply describe your project in one sentence.
- **No Clutter:** Receive a clean, well-structured project scaffold.
- **Automated File Creation:** StartGen builds directories and files for you—no manual setup needed.

## Installation & Setup

1. **Download the Executable:**  
   Grab the `startgen` binary for your system.
   
2. **Set Environment Variables (if required):**  
   ```bash
   export OPENAI_API_KEY=your_openai_api_key_here
   export DEVSEARCH_API_KEY=your_devsearch_api_key_here
   ```

3. **Make It Executable:**  
   ```bash
   chmod +x ./startgen
   ```

4. **Check the Help Menu:**  
   ```bash
   ./startgen --help
   ```

## Basic Usage

Run StartGen with a descriptive prompt:
```bash
./startgen "A Django-based blog with user authentication and a commenting system"
```

**What Happens Next?**  
StartGen interprets your prompt and generates a structured project skeleton. By default, it saves files to `./startgen_output`, but you can specify another directory if you’d like.

## Options

- `--output-dir <path>`: Specify a custom output directory.  
  Example:
  ```bash
  ./startgen "A Node.js CLI tool that converts CSV to JSON" --output-dir ./my_new_project
  ```

## Output

Upon completion, you’ll see a JSON response with two sections:

- **project_structure:** A nested JSON structure representing directories and files.
- **boilerplate_code:** Mapped code snippets or starter code for each file.

StartGen will then create those directories and files right in your chosen output directory. Easy-peasy!

## Example Workflow

1. **Think of a project:**  
   “I want a React frontend that fetches data from a public API and displays it in a grid.”
   
2. **Run StartGen:**  
   ```bash
   ./startgen "A React frontend that fetches data from a public API and displays it in a grid"
   ```
   
3. **Check the Output Directory:**  
   Voilà! Your new React project structure is ready and waiting.

## Fun Prompt

Feeling adventurous?
```bash
./startgen "A Python script that helps my cat tweet motivational quotes every morning"
```
Result? A perfectly valid project layout for your feline social media assistant!

## Troubleshooting Tips

- **No Output or Errors?**  
  Check your environment variables and API keys.
  
- **Weird File Layouts?**  
  Make sure your prompt clearly describes what you want.
  
- **Not Quite Right?**  
  Tweak your prompt and try again. You can refine until you get the perfect boilerplate.

## Support

For questions, issues, or feedback, visit our GitHub repository and open an issue. We’re always looking to make StartGen even more magical!

---

**Happy coding!** Go forth and let StartGen kickstart your next amazing project.  
