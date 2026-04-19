# ============================================================
# Stage 1: Builder
# ============================================================
FROM golang:1.25.3-alpine AS builder

RUN apk add --no-cache git curl

# Install templ CLI
RUN go install github.com/a-h/templ/cmd/templ@latest

WORKDIR /build

# Layer-cache dependencies before copying source
COPY go.mod go.sum ./
RUN go mod download

# Copy full source
COPY . .

# Generate templ files
RUN templ generate

# Build binary — strip debug info for smaller image
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /portfolio \
    ./cmd/server

# ============================================================
# Stage 2: Final image
# ============================================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

# Non-root user
RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

# Copy binary
COPY --from=builder /portfolio .

# Copy static assets and content (needed at runtime)
COPY --from=builder /build/web/static ./web/static
COPY --from=builder /build/content    ./content

USER app

EXPOSE 8080

CMD ["./portfolio"]