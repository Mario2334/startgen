from app import app

def test_predict():
    client = app.test_client()
    response = client.post('/predict', json={'data': 'sample'})
    assert response.status_code == 200
