from flask import Flask
from .routes import main
from .auth import auth

app = Flask(__name__)
app.register_blueprint(main)
app.register_blueprint(auth)
