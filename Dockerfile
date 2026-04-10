FROM golang:1.26-alpine

# Install build dependencies for CGO
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build with CGO enabled
RUN CGO_ENABLED=1 go build -o review-guess ./cmd/review-guess

# Create non-root user
RUN adduser -D -u 1000 app && \
    mkdir -p /app/data && \
    chown -R app:app /app

USER app

# Expose port
EXPOSE 8080

# Run the application
CMD ["./review-guess"]
