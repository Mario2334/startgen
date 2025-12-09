import pandas as pd

def convert_csv_to_json(input_file, output_file):
    """Convert a CSV file to a JSON file."""
    try:
        df = pd.read_csv(input_file)
        df.to_json(output_file, orient='records', indent=4)
        print(f'Successfully converted {input_file} to {output_file}')
    except Exception as e:
        print(f'Error: {e}')