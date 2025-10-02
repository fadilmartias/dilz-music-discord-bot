# --- Stage base (common deps) ---
    FROM golang:1.24 AS base
    WORKDIR /app
    
    # Copy go.mod & go.sum dulu (biar cache dependency)
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy source code
    COPY . .
    
    # --- Stage dev (hot reload dengan air) ---
    FROM base AS dev
    RUN go install github.com/air-verse/air@v1.62.0
    # Biar PATH-nya ke binary air (Go bin ada di /go/bin)
    ENV PATH="/go/bin:${PATH}"
    ENTRYPOINT [ "air", "-c", ".air.toml" ]
    
    # --- Stage build (compile binary untuk prod) ---
    FROM base AS builder
    RUN CGO_ENABLED=0 GOOS=linux go build -o bot ./cmd/bot
    
    # --- Stage prod (runtime ringan) ---
    FROM debian:bullseye-slim AS prod
    
    RUN apt-get update && \
        apt-get install -y ffmpeg python3 python3-pip ca-certificates && \
        pip3 install -U yt-dlp && \
        apt-get clean && rm -rf /var/lib/apt/lists/*
    
    WORKDIR /app
    
    COPY --from=builder /app/bot .
    
    CMD ["./bot"]