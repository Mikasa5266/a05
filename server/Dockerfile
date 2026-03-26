# syntax=docker/dockerfile:1

# Build stage: compile static binary to reduce runtime image size.
FROM golang:1.25-alpine AS builder
WORKDIR /src

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/app main.go

# Runtime stage: minimal image containing only binary and runtime assets.
FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache \
    ca-certificates \
    poppler-utils \
    tesseract-ocr \
    tesseract-ocr-data-chi_sim \
    tesseract-ocr-data-eng \
    tzdata \
 && mkdir -p /app/uploads

ENV TESSDATA_PREFIX=/usr/share/tessdata

COPY --from=builder /out/app /app/app
COPY config.yaml /app/config.yaml

EXPOSE 8080

CMD ["/app/app"]
