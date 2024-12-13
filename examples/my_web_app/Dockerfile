FROM python:3.8

WORKDIR /app

COPY backend/requirements.txt .
RUN pip install -r requirements.txt

COPY ./backend ./backend

EXPOSE 5000

CMD [ "python", "./backend/app/app.py" ]