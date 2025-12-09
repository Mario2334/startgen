from setuptools import setup, find_packages

setup(
    name='csv_to_json_tool',
    version='0.1.0',
    packages=find_packages(),
    install_requires=[
        'pandas==1.5.3'
    ],
    entry_points={
        'console_scripts': [
            'csv2json=csv_to_json.cli:main',
        ],
    },
)