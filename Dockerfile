# 使用Golang 1.18作为基础镜像
FROM golang:1.18

# 设置工作目录
WORKDIR /app

# 复制项目文件到镜像中
COPY . .

# 安装构建依赖
RUN apt-get update && apt-get install -y \
    git \
    imagemagick \
    libarchive-tools \
    curl \
    ffmpeg \
    gifsicle \
    python3 \
    python3-pip \
    && rm -rf /var/lib/apt/lists/*

# 安装Python依赖(可选)
RUN pip3 install --no-cache-dir -r requirements.txt

# 构建Go二进制文件
RUN go build -o moe-sticker-bot cmd/moe-sticker-bot/main.go

# 安装项目依赖(可选)
 RUN install tools/msb_emoji.py /usr/local/bin/ \
     && install tools/msb_kakao_decrypt.py /usr/local/bin/ \
     && install tools/msb_rlottie.py /usr/local/bin/

# 暴露端口(如果需要)
# EXPOSE 8080

# 设置入口点
CMD ["./moe-sticker-bot"]