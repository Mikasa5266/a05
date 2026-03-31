# syntax=docker/dockerfile:1

# Build stage: compile static binary to reduce runtime image size.
FROM golang:1.25-alpine AS builder
WORKDIR /src

# 优化1：替换 Alpine APK 源为国内源（解决下载 ca-certificates 等慢的问题）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
# 优化2：Go 代理配置保留（你的配置没问题，补充说明）
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/app main.go

# Runtime stage: minimal image containing only binary and runtime assets.
FROM alpine:3.20
WORKDIR /app

# 优化1：Runtime 阶段也替换 APK 源（解决下载 tesseract 等依赖慢的问题）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

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

EXPOSE 8082

CMD ["/app/app"]