# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git nodejs npm

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Install npm packages and build CSS
COPY package.json ./
RUN npm install beercss material-dynamic-colors @tailwindcss/cli
RUN mkdir -p cmd/server/static/css

# Copy source code
COPY . .

# Build Tailwind CSS
RUN cd cmd/server && npx @tailwindcss/cli -i ./static/css/input.css -o ./static/css/output.css

# Build the Go binary with embedded files (requires specifier for local Apple Silicon)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/server ./cmd/server

# Final stage - minimal image
FROM alpine:3.18

WORKDIR /app

# Copy only the built binary
COPY --from=builder /app/server /app/server

# Expose port
EXPOSE 8080

# Run the binary
CMD ["/app/server"]
