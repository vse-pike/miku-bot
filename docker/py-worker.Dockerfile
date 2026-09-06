FROM python:3.12-slim
RUN apt-get update \
    && apt-get install -y --no-install-recommends ffmpeg \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY worker/requirements.txt ./
RUN pip install --no-cache-dir -r requirements.txt
COPY worker/main.py ./

ENV HOST=0.0.0.0
EXPOSE 5005
CMD ["python", "main.py"]
