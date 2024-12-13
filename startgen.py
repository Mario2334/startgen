import argparse
import json
import re
import os
import sys
from time import sleep

from yaspin import yaspin
from haystack import Pipeline
from haystack.components.builders import PromptBuilder
from haystack.components.generators import OpenAIGenerator
from haystack.components.websearch import SerperDevWebSearch
from haystack.utils import Secret
from alive_progress import alive_bar

# project_path = os.environ.get('PROJECT_PATH')

OPENAI_API_KEY = os.environ.get('OPENAI_API_KEY')
DEVSEARCH_API_KEY = os.environ.get('DEVSEARCH_API_KEY')

if not OPENAI_API_KEY:
    print("❌ Error: The environment variable 'OPENAI_API_KEY' is not set. Please set it to use the OpenAI API.")
    sys.exit(1)  # Exit if the API key is missing

def generate_boilerplate(description):
    pipe = Pipeline()


    # Define the prompt template
    prompt_template = (
        """
        Return as JSON. 
        Include:  project_structure, boilerplate_code
        You are a software assistant that generates boilerplate code. 
        The sections (headers should be same) should include Project Structure (without comments), Boilerplate Code.
        The Project Structure should be represented as a nested json as a file system hierarchy and the file should have value as None
        Boilerplate Code file should directory mention from project path and it should be plain key-value pairs where key is file path and value is code text. 
        Validate the json and correct it"""
        "Generate boilerplate for a project described as: {{description}}"
    )
    web_search = SerperDevWebSearch(api_key=Secret.from_token(DEVSEARCH_API_KEY), top_k=2)

    pipe.add_component(instance=PromptBuilder(template=prompt_template), name="prompt_builder")
    pipe.add_component("llm", instance=OpenAIGenerator(api_key=Secret.from_token(OPENAI_API_KEY),
                                                       model='o1-preview'))
    pipe.connect("prompt_builder", "llm")
    pipe.connect("prompt_builder", "llm")

    if DEVSEARCH_API_KEY:
        pipe.add_component("search", web_search)
        main_answer = pipe.run({"search":{"query":description}, "prompt_builder": {"description": description}})
    else:
        main_answer = pipe.run({"prompt_builder": {"description": description}})

    ai_answer = main_answer.get("llm",{}).get('replies',[])[0]
    js = get_json_from_openai_response(ai_answer)

    #
    # with open('/Users/sanket/projects/startgen/startgen_output/boilerplate.txt', 'r') as fin:
    #     text = fin.read()

    # return text
    return js

def get_json_from_openai_response(text):
    try:
        json_response = json.loads(text)
    except json.decoder.JSONDecodeError:
        json_pattern = r'```json\n(.*?)\n```'

        # Find all matching JSON blocks
        json_response = re.match(json_pattern, text, re.DOTALL).group(1)
        json_response = json.loads(json_response)

    return json_response


# Function to process the directory tree and create files
def create_files_from_structure(directory_structure, parent_dir):
    for name, value in directory_structure.items():
        # Construct full path for the current directory or file
        current_path = os.path.join(parent_dir, name)

        if isinstance(value, dict):  # If the value is a directory (with files inside)
            os.makedirs(current_path, exist_ok=True)  # Create the directory
            create_files_from_structure(value, current_path)  # Recursively process nested directories
        else:  # If the value is a file (ie value is None)
            os.makedirs(os.path.dirname(current_path), exist_ok=True)  # Ensure the parent directory exists
            if os.path.isdir(current_path): os.remove(current_path)
            with open(current_path, 'w') as f:  # Create the file
                pass  # Empty file

def extract_project_structure(answer):
    client = OpenAIGenerator(model="gpt-4o-mini", api_key=Secret.from_token(OPENAI_API_KEY))
    template = f"""
    Return as JSON.
    Convert project structure to json map from this query, only write code:
    {answer}
    """
    response = client.run(template)
    json_text = response['replies'][0]
    data_dict = get_json_from_openai_response(json_text)
    # Parse the matched JSON string into a Python dictionary
    return data_dict

def write_file_from_response(path, text, base_dir):
    file_path = os.path.join(base_dir, path)
    with open(file_path, 'w') as f:
        f.write(text)


def main():
    parser = argparse.ArgumentParser(description="StartGen - AI-powered boilerplate generator.")
    parser.add_argument("prompt", type=str, help="Describe your project in plain English.")
    parser.add_argument("--output-dir", type=str, required=False  ,help="Directory to save the generated boilerplate.")

    args = parser.parse_args()

    output_dir= args.output_dir if args.output_dir else os.getcwd()

    print("🔄 Processing your request...")

    with alive_bar(5, title="Initializing components") as bar:
        for _ in range(5):
            sleep(0.5)  # Simulate work
            bar()  # Increment progress

    with yaspin(text="Generating boilerplate...", color="cyan") as spinner:
        try:
            boilerplate = generate_boilerplate(args.prompt)
            if boilerplate:
                spinner.text = "Boilerplate generated successfully!"
                spinner.ok("✔")  # Spinner success
            else:
                spinner.text = "Failed to generate boilerplate."
                spinner.fail("✘")  # Spinner failure
        except Exception as e:
            spinner.text = f"Error: {e}"
            spinner.fail("✘")

    if boilerplate:
        print("\n✅ Boilerplate generation successful! Here are the details:\n")
        print(boilerplate)

        code_struct = boilerplate['project_structure']
        code_files = boilerplate['boilerplate_code']

        ## Process Project Structure
        # extracted_project_struct = extract_project_structure(code_struct)
        # print(json.dumps(extracted_project_struct, indent=4))
        create_files_from_structure(code_struct, output_dir)

        for code_file in code_files.keys():
            write_file_from_response(code_file, code_files[code_file], output_dir)

        print("🚀 Project files have been generated successfully!")

    else:
        print("❌ Boilerplate generation failed! Please check your input and try again.")



if __name__ == "__main__":
    main()