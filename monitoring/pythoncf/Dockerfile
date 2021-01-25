FROM python:3.8

COPY requirements.txt .
RUN pip3 install -U -r requirements.txt

COPY *.py ./

CMD ["python3", "./main.py"]
