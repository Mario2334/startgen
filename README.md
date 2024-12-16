
# 🚀 Welcome to StartGen! 🌟

✨ **StartGen** is your AI-powered buddy that brings your project ideas to life—effortlessly! Tired of starting from scratch? Just describe your idea in plain English, and let StartGen handle the rest. No fuss, no hassle. 💡

---

## What Can StartGen Do for You? 🛠️

- **👨‍💻 Plain English Prompts:** Simply describe your project in one sentence, and voila!
- **🎯 No Clutter:** Get a clean, well-organized project scaffold.
- **✨ Instant Files:** Automatic directory and file creation—ready for your magic touch.

---

## 🚧 Installation & Setup 🛠️

### Using Brew

#### Install From brew
```bash
   brew tap Mario2334/startgen
   brew install startgen
```

### Using Binary
1. **🔽 Download the Executable:**  
   - Grab the `startgen` binary for your system.

2. **🔑 Set Your Environment Variables:**  
   ```bash
   export OPENAI_API_KEY=your_openai_api_key_here
   export PROJECT_PATH=your_project_directory
   ```

3. **🏗️ Make It Executable:**  
   ```bash
   chmod +x ./startgen
   ```

4. **📚 Need Help?**  
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
🧠 StartGen uses your description to create:
- A **React frontend** with components and pages.
- A **Flask backend** with REST API boilerplate.
- Database setup scripts for **PostgreSQL**.

---

## Options 🛠️

- **`-o <path>`:**  
  Specify where your project files will live.  
  Example:
  ```bash
  ./startgen startgen -o "<dir>" "A Node.js CLI tool that converts CSV to JSON"
  ```

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
