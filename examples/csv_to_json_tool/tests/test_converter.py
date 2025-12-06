import unittest
import pandas as pd
from csv_to_json.converter import convert_csv_to_json
import os

class TestConverter(unittest.TestCase):

    def setUp(self):
        self.input_csv = 'test.csv'
        self.output_json = 'test.json'
        # create a sample CSV file
        data = {'name': ['Alice', 'Bob'], 'age': [25, 30]}
        df = pd.DataFrame(data)
        df.to_csv(self.input_csv, index=False)

    def tearDown(self):
        os.remove(self.input_csv)
        if os.path.exists(self.output_json):
            os.remove(self.output_json)

    def test_convert_csv_to_json(self):
        convert_csv_to_json(self.input_csv, self.output_json)
        self.assertTrue(os.path.exists(self.output_json))

if __name__ == '__main__':
    unittest.main()