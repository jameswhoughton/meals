# Stage 1: Build Tailwind CSS using Node
FROM node:20 AS frontend-builder

WORKDIR /app

COPY web/package*.json ./
RUN npm install

COPY web/ ./
RUN npx tailwindcss -i ./src/input.css -o ./static/main.css --minify

# Stage 2: Build Go binary
FROM golang:1.23 AS go-builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Copy built frontend into appropriate location
COPY --from=frontend-builder /app/static ./web/static

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/server

# Stage 3: Final runtime image
FROM alpine:latest

# Create user and group
RUN addgroup -S www && adduser -S www -G www

WORKDIR /app

COPY --from=go-builder /app/app .

RUN chown -R www:www /app

USER www

EXPOSE 8000

CMD ["./app"]

