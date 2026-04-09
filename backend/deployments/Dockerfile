FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装依赖
RUN apk add --no-cache git

# 设置 Go 模块代理（使用国内镜像加速）
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

# 复制 go.mod
COPY go.mod ./

# 下载依赖（如果没有 go.sum，会自动生成）
RUN go mod download && go mod tidy

# 复制源代码
COPY . .

# 重新下载并整理依赖
RUN go mod download

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server cmd/server/main.go

# 最终镜像
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# 从builder复制二进制文件
COPY --from=builder /app/server .

# 暴露端口
EXPOSE 8080

# 运行
CMD ["./server"]
