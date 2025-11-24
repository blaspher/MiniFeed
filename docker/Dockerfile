# 使用 Go 1.23 Alpine 镜像作为基础镜像
FROM golang:1.23-alpine

# 设置工作目录
WORKDIR /app

# 将整个项目（包括 minifeed 文件夹中的代码）复制到容器的 /app 目录
COPY . .

# 安装 Go 依赖
RUN go mod tidy

# 构建 Go 应用
# 在根目录构建 cmd/server 下的主程序
RUN go build -o main ./cmd/server

# 暴露应用端口
EXPOSE 8888

# 启动应用
CMD ["./main"]
