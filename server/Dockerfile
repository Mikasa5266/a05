# syntax=docker/dockerfile:1

# --- 阶段 1: 编译阶段 ---
FROM golang:1.25-alpine AS builder
WORKDIR /src

# 使用国内源加速
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache ca-certificates tzdata

# 处理 Go 依赖
COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn
RUN go mod download

# 编译二进制文件
COPY . .

# 关键修复：在编译前强制整理依赖，解决 go.mod 报错
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/server

# --- 阶段 2: 运行阶段 ---
FROM alpine:3.20
WORKDIR /app

# 安装运行时依赖、配置国内源、创建普通用户
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache \
    ca-certificates \
    poppler-utils \
    tesseract-ocr \
    tesseract-ocr-data-chi_sim \
    tesseract-ocr-data-eng \
    tzdata && \
    adduser -D appuser && \
    mkdir -p /app/uploads && \
    chown appuser:appuser /app/uploads

ENV TESSDATA_PREFIX=/usr/share/tessdata

# 拷贝编译好的程序和配置文件
COPY --from=builder /out/app /app/app
COPY config.yaml /app/config.yaml

# 使用普通用户运行
USER appuser

EXPOSE 8082

CMD ["/app/app"]