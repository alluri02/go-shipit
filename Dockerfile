# Multi-stage Dockerfile for go-shipit
#
# Multi-stage builds = small final image (just the binary, no Go toolchain).
# Final image is ~10MB instead of ~1GB.
#
# C# equivalent: multi-stage with dotnet/sdk → dotnet/aspnet
# Java equivalent: multi-stage with maven → eclipse-temurin (JRE only)

# ─── Stage 1: Build ───
# Use the Go image to compile our binary
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependency files first (better layer caching)
# If go.mod/go.sum haven't changed, Docker reuses the cached layer
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o shipit ./cmd/shipit

# ─── Stage 2: Run ───
# Use a minimal base image — just the binary, nothing else
# scratch = empty image (0 bytes). Our Go binary is statically linked.
#
# C# equivalent: FROM mcr.microsoft.com/dotnet/aspnet:8.0-alpine (still ~100MB)
# Java equivalent: FROM eclipse-temurin:21-jre-alpine (still ~150MB)
# Go: FROM scratch (0 bytes + our binary ≈ 10MB total)
FROM scratch

# Copy the binary from the builder stage
COPY --from=builder /app/shipit /shipit

# Copy CA certificates for HTTPS calls (AI API, Azure, etc.)
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose the default port
EXPOSE 8080

# Run the binary
ENTRYPOINT ["/shipit"]
CMD ["serve"]
