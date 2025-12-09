
# 🚀 Welcome to StartGen! 🌟

✨ **StartGen** is your AI-powered buddy that brings your project ideas to life—effortlessly! Tired of starting from scratch? Just describe your idea in plain English, and let StartGen handle the rest. No fuss, no hassle. 💡

---

## What Can StartGen Do for You? 🛠️

- **👨‍💻 Plain English Prompts:** Simply describe your project in one sentence, and voila!
- **🧠 Smart Multi-Stage Generation:** StartGen thinks, plans, and then generates your project in three intelligent stages
- **🎯 No Clutter:** Get a clean, well-organized project scaffold.
- **✨ Instant Files:** Automatic directory and file creation—ready for your magic touch.
- **📊 Generation Insights:** See what technologies, patterns, and components were identified for your project

---

## 🚧 Installation & Setup 🛠️

**🔑 Set Your Environment Variables:**  
   ```bash
   export OPENAI_API_KEY=your_openai_api_key_here
   ```

### Using Brew

#### Install From brew
```bash
   brew tap Mario2334/startgen
   brew install startgen
```

### Using Binary
1. **🔽 Download the Executable:**  
   - Grab the `startgen` binary for your system.

2. **🏗️ Make It Executable:**  
   ```bash
   chmod +x ./startgen
   ```

3. **📚 Need Help?**  
   ```bash
   ./startgen --help
   ```

### Building from Source

If you want to build and run StartGen directly from source:

1. **📋 Prerequisites:**  
   - Go 1.21 or higher installed
   - Git

2. **📥 Clone the Repository:**  
   ```bash
   git clone https://github.com/Mario2334/startgen.git
   cd startgen
   ```

3. **🔨 Build the Project:**  
   ```bash
   go build -o startgen .
   ```

4. **🚀 Run StartGen:**  
   ```bash
   ./startgen "Your project description here"
   ```

   Or run directly without building:
   ```bash
   go run . "Your project description here"
   ```

5. **📚 View Options:**  
   ```bash
   ./startgen --help
   ```

---

## ⚙️ How to Use StartGen 🎉

Run StartGen with a simple, fun prompt:

```bash
startgen -o "<dir>" "A React app with a Flask backend and PostgreSQL database."
```

### **What Happens Next?**  
🧠 StartGen uses a smart three-stage process:

1. **🤔 Thinking Stage:** Analyzes your requirements and identifies technologies, architecture patterns, and key components
2. **📋 Planning Stage:** Creates a detailed implementation plan with directory structure and file descriptions
3. **⚡ Generation Stage:** Generates all the code based on the thoughtful plan

For your React + Flask + PostgreSQL app, you'll get:
- A **React frontend** with components and pages.
- A **Flask backend** with REST API boilerplate.
- Database setup scripts for **PostgreSQL**.
- **Metadata** showing the technologies and architecture decisions made

---

## Options 🛠️

- **`-o <path>`:**  
  Specify where your project files will live.  
  Example:
  ```bash
  ./startgen -o "<dir>" "A Node.js CLI tool that converts CSV to JSON"
  ```

- **`--save-metadata`:**  
  Save detailed generation metadata (thinking and planning outputs) to `.startgen-metadata.json` in your project.  
  Example:
  ```bash
  ./startgen --save-metadata "A Python web scraper"
  ```

---

## 🧠 How Multi-Stage Generation Works �

StartGen uses a sophisticated three-stage approach to ensure high-quality, well-structured projects:

### Stage 1: Thinking 🤔
The AI analyzes your project description to identify:
- Technologies and frameworks needed
- Appropriate architecture patterns
- Key components and their relationships
- Assumptions about unclear requirements
- Technical constraints

### Stage 2: Planning 📋
Based on the analysis, the AI creates:
- Complete directory structure
- Detailed file descriptions
- Required dependencies with versions
- Configuration file specifications

### Stage 3: Generation ⚡
Finally, the AI generates:
- All source code files
- Configuration files
- Documentation
- Setup scripts

Each stage builds on the previous one, ensuring your project is thoughtfully designed before any code is written!

---

## 🚀 Example Workflow: Bring Your Ideas to Life! 🌈

1. **💡 Dream Up a Project:**  
   *“I want a Python-based data pipeline to process weather data!”*

2. **🤖 Run StartGen:**  
   ```bash
   startgen "A Python-based data pipeline to process weather data"
   ```

3. **📂 Check Your Project Files:**  
   Your new project is ready for you to start coding. Easy-peasy lemon squeezy! 🍋

4. **📊 Review the Generation Summary:**  
   StartGen shows you what it created:
   - Technologies identified (Python, Flask, SQLAlchemy, etc.)
   - Architecture pattern used (MVC, microservices, etc.)
   - Key components included

---

## Need a Laugh? 😄

Try a silly prompt:
```bash
startgen "A Python script to help my dog learn to code."
```
🐶 You’ll still get a valid project scaffold… paw-some!

---

## Troubleshooting Tips 🛠️

- **No Output or Errors?**  
  👉 Check your environment variables (`OPENAI_API_KEY` and `PROJECT_PATH`).

- **Weird Layouts?**  
  📝 Make sure your prompt is clear and descriptive.

- **Not What You Expected?**  
  ✍️ Refine your prompt and try again!

---

## Support & Feedback ❤️

💌 Got questions or suggestions? Open an issue on our [GitHub repository](https://github.com/your-repo). We’d love to hear from you!

---

✨ **StartGen**: Where ideas meet reality, one project at a time. Happy coding! 🎉
