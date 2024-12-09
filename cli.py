import argparse
import json
import re
import os

from haystack import Pipeline
from haystack.components.builders import PromptBuilder
from haystack.components.generators import OpenAIGenerator
from haystack.components.websearch import SerperDevWebSearch
from haystack.utils import Secret


project_path = os.environ.get('PROJECT_PATH')

OPENAI_API_KEY = os.environ.get('OPENAI_API_KEY')
DEVSEARCH_API_KEY = os.environ.get('DEVSEARCH_API_KEY')

if project_path is None:
    project_path = os.getcwd()

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
        Boilerplate Code file should directory mention from project path.
        Validate the json and correct it"""
        "Generate boilerplate for a project described as: {{description}}"
    )
    web_search = SerperDevWebSearch(api_key=Secret.from_token(DEVSEARCH_API_KEY), top_k=2)

    pipe.add_component(instance=PromptBuilder(template=prompt_template), name="prompt_builder")
    pipe.add_component("llm", instance=OpenAIGenerator(api_key=Secret.from_token(OPENAI_API_KEY),
                                                       model='gpt-4o'))
    pipe.connect("prompt_builder", "llm")
    pipe.connect("prompt_builder", "llm")
    pipe.add_component("search", web_search)

    # Generate the boilerplate code
    main_answer = pipe.run({"search":{"query":description}, "prompt_builder": {"description": description}})
    main_answer = main_answer.get("llm",{}).get('replies',[])[0]
    js = get_json_from_openai_response(main_answer)

    #
    # with open('/Users/sanket/projects/startgen/startgen_output/boilerplate.txt', 'r') as fin:
    #     text = fin.read()

    # return text
    return js

def get_json_from_openai_response(text):
    json_response = None
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
            # os.makedirs(os.path.dirname(current_path), exist_ok=True)  # Ensure the parent directory exists
            with open(current_path, 'w') as f:  # Create the file
                pass  # Empty file

def extract_project_structure(answer):
    client = OpenAIGenerator(model="gpt-3.5-turbo", api_key=Secret.from_token(OPENAI_API_KEY))
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

def write_file_from_response(path, text):
    file_path = os.path.join(project_path, path)
    with open(file_path, 'w') as f:
        f.write(text)


def main():
    parser = argparse.ArgumentParser(description="StartGen - AI-powered boilerplate generator.")
    parser.add_argument("prompt", type=str, help="Describe your project in plain English.")
    parser.add_argument("--output-dir", type=str, default="./startgen_output", help="Directory to save the generated boilerplate.")

    args = parser.parse_args()

    print("Processing your request...")
    boilerplate = generate_boilerplate(args.prompt)

    if boilerplate:
        print("\nGenerated Boilerplate:\n")
        print(boilerplate)

        code_struct = boilerplate['project_structure']
        code_files = boilerplate['boilerplate_code']

        ## Process Project Structure
        # extracted_project_struct = extract_project_structure(code_struct)
        # print(json.dumps(extracted_project_struct, indent=4))
        create_files_from_structure(code_struct, project_path)

        for code_file in code_files.keys():
            write_file_from_response(code_file, code_files[code_file])




if __name__ == "__main__":
    main()