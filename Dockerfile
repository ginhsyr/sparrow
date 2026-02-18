# Stage 1: Build the binary
FROM golang:1.24.3-bookworm AS builder

# Install build tools (adjust for your language)
RUN apt-get update && \
    apt-get install -y build-essential git ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o Sparrow main.go

# Stage 2: Runtime image
FROM ubuntu:22.04

# Install runtime deps
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /workplace/run
RUN useradd -ms /bin/bash hsyr

# Copy files from builder
COPY .env ./
COPY --from=builder /app/Sparrow .

# Set permissions
RUN chown hsyr:hsyr Sparrow && \
    chmod +x Sparrow

# Run as non-root user
USER hsyr
CMD ["./Sparrow"]
