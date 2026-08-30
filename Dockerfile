FROM golang:1.24-bookworm

RUN apt-get update && \
    apt-get install -y --no-install-recommends ffmpeg python3 python3-venv ca-certificates curl unzip nodejs npm && \
    rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://deno.land/install.sh | sh -s -- -y && \
    ln -s /root/.deno/bin/deno /usr/local/bin/deno

RUN python3 -m venv /tmp/yt-dlp-venv && \
    /tmp/yt-dlp-venv/bin/pip install --no-cache-dir -U yt-dlp

ENV PATH="/tmp/yt-dlp-venv/bin:${PATH}"

WORKDIR /app

COPY server/go.mod server/go.sum ./server/
RUN cd server && go mod download

COPY . .

RUN cd server && go build -o /app/nt-music-server .

EXPOSE 8765

CMD ["sh", "-c", "cd /app/server && exec /app/nt-music-server"]
