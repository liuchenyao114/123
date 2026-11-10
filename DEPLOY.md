# 部署指南

> 适用版本 `v0.5.1`。选择平台的完整推理见本文「一、怎么选」。

## 一、怎么选

这个项目的两个特性决定了平台选择，比「哪个免费额度大」更重要：

1. **cmcc 通道依赖常驻的 WebSocket 长连接**（心跳、入站消息自动记录接收人）。平台一旦休眠，连接断开，休眠期间收不到入站消息。
2. **默认用 SQLite**，用户、通道配置（API Key、接收人）、消息历史都在本地文件里。平台若没有持久磁盘，每次重启 / 重新部署都要重新配置一遍。

| 平台 | 信用卡 | 休眠 | 持久磁盘 | 结论 |
| --- | --- | --- | --- | --- |
| **Serv00 / ct8.pl** | 不需要 | 不休眠 | 3 GB | ✅ **正式部署首选** |
| Hugging Face Spaces | 不需要 | 约 48 小时闲置 | 无（要付费） | 备选，资源足、休眠窗口长 |
| Render | 不需要 | 15 分钟无流量 | 无 | ⚠️ 只适合临时演示 |
| Fly.io / Koyeb / Oracle / GCP / AWS / Azure | 需要 | — | — | ❌ 不符合「不用信用卡」 |

**结论：要长期跑，选 Serv00。** 它是唯一同时满足「不用信用卡 + 进程常驻 + 有持久磁盘 + 原生支持 Go」的免费选项。Render 是上一轮按通用后端给的推荐，按本项目实际用途看只能当临时演示，这里予以修正。

**Serv00 的代价**（务必先接受再动手）：

- 系统是 **FreeBSD，不是 Linux**。仓库 Release 里的 `message-pusher-linux-amd64` 是 Linux ELF，**不能直接用**，必须用源码在这台机器上重新编译（见第二节）。
- 512 MB 内存 / 3 GB 磁盘 / 最多开 3 个端口 / 不支持 Docker。
- **每 90 天必须登录一次面板或 SSH**，否则账号连同数据一起被回收且无法恢复。请设日历提醒。
- 注册名额经常满，需要等新服务器开放；禁止代理/VPN/隧道类用途。
- 服务器在波兰，国内访问延迟较高。

> 如果已经有自己的服务器或 VPS，直接用 `v0.5.1` 的二进制跑是稳定得多的选择：不休眠、有磁盘、无保号规则。下面两套方案都是「没有自己机器」时的替代。

## 二、方案 A：Serv00（推荐，长期运行）

1. 在 serv00.com 用邮箱注册（不要信用卡、不要实名），记下面板地址、SSH 地址、分配到的端口。
2. 在 DevilWEB 面板里申请一个 TCP 端口，并把免费二级域名（或自己的域名）代理到该端口；同时开启 SSL。面板具体选项以实际界面为准。
3. SSH 登录后准备源码。**不要下载 Release 二进制**，用源码编译：

```sh
mkdir -p ~/apps && cd ~/apps
# 方式 1：从 GitHub 拉源码（若网络不通，改用方式 2）
git clone --depth 1 https://github.com/liuchenyao114/123.git message-pusher
cd message-pusher
# 方式 2：本地下载仓库 zip 后用 scp 上传，再解压到 ~/apps/message-pusher

# 仓库已附带 web/build，go:embed 需要它，因此不需要装 Node
CGO_ENABLED=1 go build -ldflags "-s -w" -o message-pusher .
```

> 若 `go build` 因 cgo / go-sqlite3 报缺 C 编译器，改用面板自带的 Go 版本或在编译命令前指定 `CC=clang`；FreeBSD 一般自带 clang。

4. 建立数据目录并启动：

```sh
mkdir -p ~/data
export SESSION_SECRET="$(head -c 32 /dev/urandom | od -An -tx1 | tr -d ' \n')"
export PORT=<面板分配给你的端口>
export SQLITE_PATH="$HOME/data/message-pusher.db"
nohup ./message-pusher > ~/data/message-pusher.log 2>&1 &
```

5. 用面板的 Cron / 常驻进程功能配置开机与异常退出后自动拉起（例如每 5 分钟检查进程是否存在）。
6. 浏览器打开你的域名，用 `root` / `123456` 登录并**立刻改密**；再按下方「四、部署后检查清单」配置。

**90 天保号**：在日历里设一个季度提醒，登录一次面板或 SSH 即可。

## 三、方案 B：Render（最快上线，仅临时演示）

仓库根目录已提供 `render.yaml`（Blueprint）。在 Render 控制台 → Blueprints → 连接本仓库 → 一键创建 Free Web Service，或直接新建 Web Service 选 Docker、关联 `main` 分支。

平台自动注入 `PORT`（默认 10000），程序会读取它，无需改启动命令。Blueprint 会自动生成 `SESSION_SECRET`。

**必须接受的限制**：

| 限制 | 影响 | 缓解 |
| --- | --- | --- |
| 15 分钟无流量休眠，冷启动 30–60 秒 | cmcc 长连接断开，休眠期间入站消息丢失 | 用 UptimeRobot 之类每 10 分钟 ping 一次保活（Render 不鼓励）；或只把「默认接收人」写死在通道配置里，不依赖入站消息 |
| 免费层无持久磁盘 | 每次部署 / 重启后 SQLite 里的用户、通道、消息全部清空 | 设 `SQL_DSN` 指向外部 MySQL；或只当一次性演示 |
| 出站 25/465/587 被封 | 邮件通道发不出去 | 改用其他通道 |
| 免费 Postgres 30 天过期 | **对本项无用**：项目只有 mysql / sqlite 驱动，连不上 Postgres | 不要创建 Render Postgres |

想少休眠一点，可以改用 **Hugging Face Spaces**（Docker SDK，无需信用卡，免费 2 vCPU / 16 GB，约 48 小时才休眠）：把 `Dockerfile` 的 `EXPOSE` 和启动端口改成 **7860**，设同样的三个环境变量即可。它的容器同样无状态，SQLite 仍会丢。

## 四、环境变量速查

| 变量 | 说明 |
| --- | --- |
| `PORT` | 监听端口，程序原生读取；不设则用 `--port` 参数（默认 3000） |
| `SESSION_SECRET` | **必设**，会话签名密钥；更换后需重新登录 |
| `SQLITE_PATH` | SQLite 文件路径，默认写在进程工作目录（Dockerfile 里是 `/data`）。务必指向持久目录 |
| `SQL_DSN` | 设置后改用 MySQL，例如 `user:pass@tcp(host:3306)/message_pusher?charset=utf8mb4&parseTime=True&loc=Local`。用它可以绕开无持久磁盘的问题 |
| `TZ` | 时区，建议 `Asia/Shanghai` |

## 五、部署后检查清单

1. 立刻修改默认密码 `root` / `123456`。
2. 系统设置里关闭「允许新用户注册」，配置服务器地址。
3. 推送设置里设置**全局鉴权令牌**；注意 `v0.5.1` 起鉴权收紧，不带令牌的旧 `/push` 调用会返回 401。
4. 新建通道：类型选「中国移动新消息（5G 消息）」，填 API Key 和默认接收人；或配置企业微信 / 飞书 / 钉钉等。
5. 老数据库执行 `DELETE FROM channels WHERE type IN ('tencent_alarm','multi');` 清理失效通道。
6. 发一条测试推送：

```bash
curl "http://<你的地址>/push/root?title=测试&description=hello&token=<全局令牌>"
```

7. 定期备份 `SQLITE_PATH` 指向的 db 文件（Serv00 有每日自动备份，但仍建议自己留一份）。
