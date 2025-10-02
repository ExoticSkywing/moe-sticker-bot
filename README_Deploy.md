## 非容器化部署与常见坑位修复指引（moe-sticker-bot）

本指引面向需要改动代码、在裸机/虚机上直接部署（非容器化）的场景，覆盖系统依赖、视频/动态贴纸转换依赖、WebApp 配置、systemd 守护与常见问题自检。

### 一、关键踩坑点速记
- ffmpeg 功能不全：发行版自带 ffmpeg 往往不支持 paletteuse 的 `dither=atkinson`，导致视频→GIF 失败，只留 `.webm`。
- 失效 PPA：`ppa:jonathonf/ffmpeg-4` 已不可用，`apt update` 会报错中断。
- `.tgs`→GIF 依赖缺：未装 `rlottie-python` 与 `Pillow`（`PIL`）会触发 `lottieToGIF ERROR!`。
- PATH 未包含工具：`msb_rlottie.py`、`gifsicle`、`convert`、`bsdtar`、`ffmpeg`、`python3` 必须可执行。
- 路径行为差异：
  - “整包下载 ZIP”路径会做转换（video→GIF、tgs→GIF）。
  - WebApp/WhatsApp 导出默认不转 GIF（video 保留 `.webm`，animated 保留 `.tgs`，static 保留 `.webp`）。
- 部署形态未明确：缺少 systemd 守护、日志路径、数据目录与权限示例。
- WebApp 配置不清：`--webapp_url`/`--webapp_listen_addr`/`--webapp_data_dir` 与 nginx 反代关系易混淆。

---

### 二、一次到位的最小修补与安装步骤

1) 清理失效 PPA 与刷新索引
```bash
sudo add-apt-repository -r ppa:jonathonf/ffmpeg-4 || true
sudo rm -f /etc/apt/sources.list.d/jonathonf-ubuntu-ffmpeg-4-*.list
sudo apt-get update -y
```

2) 安装系统依赖（不含 ffmpeg，后续用静态版）
```bash
sudo apt-get install -y imagemagick libarchive-tools curl gifsicle python3 python3-pil
```

3) 安装支持 atkinson 的静态版 ffmpeg（建议）
```bash
cd /usr/local/bin
sudo curl -fsSLO https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz
sudo tar -xJf ffmpeg-release-amd64-static.tar.xz --strip-components=1 --wildcards '*/ffmpeg' '*/ffprobe'
sudo chmod +x ffmpeg ffprobe
ffmpeg -version | head -n 1   # 应显示 7.x static
```
可选验证（内存紧张可跳过或缩小输入）：
```bash
ffmpeg -filters | grep -E '^ T.*paletteuse' || true
ffmpeg -v error -f lavfi -i testsrc2 -t 0.5 \
  -lavfi "split[a][b];[a]palettegen[p];[b][p]paletteuse=dither=atkinson" \
  -f null - || true
```

4) `.tgs`→GIF 依赖与脚本入 PATH
```bash
python3 -m pip install --upgrade pip setuptools wheel
python3 -m pip install rlottie-python --no-cache-dir
sudo install -m 0755 ./tools/msb_rlottie.py /usr/local/bin/msb_rlottie.py

# 验证
python3 -c "import rlottie_python, PIL, PIL.Image; print('OK')"
which msb_rlottie.py
```

---

### 三、构建与运行（后端 Go 服务）
依赖：Go 1.18+
```bash
cd /path/to/moe-sticker-bot
go build -o ./bin/moe-sticker-bot ./cmd/moe-sticker-bot

# 最小运行（仅 Bot）
./bin/moe-sticker-bot --bot_token="YOUR_BOT_TOKEN" --data_dir=/var/lib/moe-sticker-bot --log_level=info
```

建议使用 systemd 守护：
```ini
# /etc/systemd/system/moe-sticker-bot.service
[Unit]
Description=Moe Sticker Bot (Go)
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
User=YOUR_USER
Group=YOUR_USER
WorkingDirectory=/opt/moe-sticker-bot
ExecStart=/opt/moe-sticker-bot/bin/moe-sticker-bot --bot_token=YOUR_BOT_TOKEN --data_dir=/var/lib/moe-sticker-bot --log_level=info
Restart=on-failure
RestartSec=3
StandardOutput=append:/var/log/moe-sticker-bot/stdout.log
StandardError=append:/var/log/moe-sticker-bot/stderr.log
Environment=PATH=/usr/local/bin:/usr/bin:/bin

[Install]
WantedBy=multi-user.target
```
初始化：
```bash
sudo mkdir -p /opt/moe-sticker-bot/bin /var/lib/moe-sticker-bot /var/log/moe-sticker-bot
sudo cp ./bin/moe-sticker-bot /opt/moe-sticker-bot/bin/
sudo systemctl daemon-reload
sudo systemctl enable --now moe-sticker-bot
sudo systemctl status moe-sticker-bot
```

小贴士（systemd 字段与权限说明）
- User/Group：替换为实际运行用户（建议非 root）。确保该用户对 `/var/lib/moe-sticker-bot`、`/var/log/moe-sticker-bot` 有读写权限。
- WorkingDirectory：与实际部署目录保持一致（示例使用 `/opt/moe-sticker-bot`）。
- ExecStart：把 `YOUR_BOT_TOKEN` 换成真实值；如启用 WebApp，再追加 `--webapp_url/--webapp_listen_addr/--webapp_data_dir`。
- StandardOutput/StandardError：日志文件所在目录需对运行用户可写。
- Environment=PATH：建议包含 `/usr/local/bin`（静态 ffmpeg、`msb_rlottie.py` 常在此）。

更新与重启（最小中断）
```bash
cd /path/to/moe-sticker-bot
go build -o ./bin/moe-sticker-bot.new ./cmd/moe-sticker-bot
sudo install -m 0755 ./bin/moe-sticker-bot.new /opt/moe-sticker-bot/bin/moe-sticker-bot.tmp
sudo mv /opt/moe-sticker-bot/bin/moe-sticker-bot.tmp /opt/moe-sticker-bot/bin/moe-sticker-bot
sudo systemctl restart moe-sticker-bot
sudo systemctl status moe-sticker-bot
```

常见操作
```bash
# 查看实时日志
sudo journalctl -u moe-sticker-bot -f

# 平滑重启
sudo systemctl restart moe-sticker-bot

# 服务开机自启状态
systemctl is-enabled moe-sticker-bot
```

故障排查
- 服务起不来：`sudo systemctl status moe-sticker-bot` 查看最近错误；必要时 `journalctl -u moe-sticker-bot -n 200`。
- 权限问题：确认运行用户对数据/日志目录有写权限；必要时 `sudo chown -R USER:GROUP /var/lib/moe-sticker-bot /var/log/moe-sticker-bot`。
- 找不到外部工具：确认 `which ffmpeg gifsicle convert bsdtar python3 msb_rlottie.py` 均能找到；若路径在 `/usr/local/bin`，确保已加入 PATH。

---

### 四、启用 WebApp（可选）
1) 构建 React 前端（Node 18+）
```bash
cd ./web/webapp3
npm ci
npm run build
```
2) nginx 托管静态与 API 反代（后端 API 监听 127.0.0.1:8081；公开 URL 为 https://your.domain.com/webapp）
```nginx
server {
    listen 80;
    server_name your.domain.com;

    root /path/to/moe-sticker-bot/web/webapp3/build;
    index index.html;

    location / {
        try_files $uri /index.html;
    }
    location /webapp/api/ {
        proxy_pass http://127.0.0.1:8081/webapp/api/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```
3) 后端启动参数示例（加入 WebApp）
```bash
./bin/moe-sticker-bot \
  --bot_token=YOUR_BOT_TOKEN \
  --data_dir=/var/lib/moe-sticker-bot \
  --webapp_url="https://your.domain.com/webapp" \
  --webapp_listen_addr="127.0.0.1:8081" \
  --webapp_data_dir="/var/lib/moe-sticker-bot/webapp" \
  --log_level=info
```

---

### 五、数据库（可选，MariaDB）
```bash
sudo mysql -u root
CREATE DATABASE msb DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'msb_user'@'localhost' IDENTIFIED BY 'strong_password';
GRANT ALL PRIVILEGES ON msb.* TO 'msb_user'@'localhost';
FLUSH PRIVILEGES;
```
后端参数：`--db_addr=127.0.0.1:3306 --db_user=msb_user --db_pass=strong_password`

---

### 六、路径行为说明（避免误解）
- “下载整包 ZIP”路径会进行转换：
  - 视频/WEBM → GIF（依赖 ffmpeg paletteuse atkinson）
  - 动态 `.tgs` → GIF（依赖 rlottie-python + Pillow + msb_rlottie.py）
- WebApp/WhatsApp 导出默认不转 GIF：
  - Video 保留 `.webm`，Animated 保留 `.tgs`，Static 保留 `.webp`

---

### 七、快速自检清单
```bash
which ffmpeg && ffmpeg -version | head -n 1     # /usr/local/bin/ffmpeg 7.x static
which gifsicle && gifsicle --version
which convert && convert -version                # ImageMagick
which bsdtar && bsdtar --version
python3 -c "import rlottie_python, PIL, PIL.Image; print('OK')"
which msb_rlottie.py
```

### 八、常见问题
- 只出 `.webm` 不出 `.gif`：多因 ffmpeg 不支持 `dither=atkinson`。安装静态版或降级代码到 `dither=floyd_steinberg`。
- `.tgs` 转换报错 `lottieToGIF ERROR!`：安装 `rlottie-python` 与 `python3-pil`，并确保 `msb_rlottie.py` 在 PATH。
- 依赖缺失：确认上述工具均可在 PATH 中执行。


