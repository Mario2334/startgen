from flask import Blueprint

main = Blueprint('main', __name__)

@main.route('/api/data', methods=['GET'])
def get_data():
    return {'data': 'Sample Data'}
