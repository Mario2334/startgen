from flask import Blueprint

auth = Blueprint('auth', __name__)

@auth.route('/api/login', methods=['POST'])
def login():
    return {'message': 'Login successful'}
