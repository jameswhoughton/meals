# Stage 1: Build Tailwind CSS using Node
FROM node:20 AS frontend-builder

WORKDIR /app

COPY web/ ./web
COPY frontend/package*.json ./frontend/
COPY frontend/tailwind.config.js ./frontend/

WORKDIR /app/frontend

RUN npm install && npx tailwindcss -i ../web/src/input.css -o ../web/static/main.css --minify

# Stage 2: Build Go binary
FROM golang:1.23 AS go-builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Copy built frontend into appropriate location
COPY --from=frontend-builder /app/web/static ./web/static

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/server

# Stage 3: Final runtime image
FROM scratch

COPY --from=go-builder /app/app .

EXPOSE 8000

CMD ["./app"]

