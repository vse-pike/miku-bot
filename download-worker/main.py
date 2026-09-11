import logging
import os
from urllib.parse import urlparse

from flask import Flask, jsonify, request
from yt_dlp import YoutubeDL

DOWNLOAD_DIR = os.environ.get("DOWNLOAD_DIR", "./downloads")
HOST = os.environ.get("HOST", "127.0.0.1")
PORT = int(os.environ.get("PORT", 5005))
# Нужен только для закрытого контента (18+, приватные аккаунты) — TikTok на
# таких видео требует логин. Обычные ролики качаются и без файла.
COOKIES_FILE = os.environ.get("COOKIES_FILE")

os.makedirs(DOWNLOAD_DIR, exist_ok=True)

logging.basicConfig(level=logging.INFO, format="%(asctime)s %(levelname)s %(message)s")
logger = logging.getLogger("miku-worker")

app = Flask(__name__)

# Часть сайтов (короткие клипы TikTok/Insta/X) обычно отдают уже готовый
# комбинированный файл — не нужен отдельный мёрж видео+аудио через ffmpeg,
# как для YouTube в высоком качестве. Разделение форматов — из твоего же
# MikuVL/yt_downloader/handlers/downloader.py, тут просто без async-обвязки.
LIGHT_DOMAINS = ["tiktok.com", "instagram.com", "x.com", "twitter.com"]

common_opts = {
    "merge_output_format": "mp4",
    "outtmpl": os.path.join(DOWNLOAD_DIR, "%(id)s.%(ext)s"),
    "noplaylist": True,
}
if COOKIES_FILE:
    common_opts["cookiefile"] = COOKIES_FILE

ydl = YoutubeDL({
    **common_opts,
    "format": "bestvideo[height<=1080]+bestaudio/best[height<=1080]",
})

ydl_light = YoutubeDL({**common_opts, "format": "best"})


def is_light_domain(url: str) -> bool:
    hostname = urlparse(url).hostname or ""
    return any(hostname.endswith(domain) for domain in LIGHT_DOMAINS)


@app.route("/download", methods=["POST"])
def download():
    data = request.get_json(silent=True) or {}
    url = data.get("url")
    if not url:
        return jsonify({"result": False, "description": "url is required"}), 400

    client = ydl_light if is_light_domain(url) else ydl
    logger.info("downloading %s (light=%s)", url, client is ydl_light)

    try:
        info = client.extract_info(url, download=True)
        path = client.prepare_filename(info)
    except Exception as e:
        logger.error("download failed for %s: %s", url, e)
        return jsonify({"result": False, "description": str(e)}), 500

    return (
        jsonify(
            {
                "result": True,
                "description": "",
                "path": os.path.abspath(path),
                "size": os.path.getsize(path),
                "title": info.get("title", ""),
            }
        ),
        200,
    )


if __name__ == "__main__":
    # Один процесс, без reloader — держит yt-dlp/YoutubeDL заимпортированным
    # один раз, ровно то, ради чего мы вообще уходили от os/exec-подхода.
    app.run(host=HOST, port=PORT)
