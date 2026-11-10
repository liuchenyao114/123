<p align="center">
  <a href="https://github.com/songquanpeng/message-pusher"><img src="https://raw.githubusercontent.com/songquanpeng/message-pusher/master/web/public/logo.png" width="150" height="150" alt="message-pusher logo"></a>
</p>

<div align="center">

# 消息推送服务

_✨ 搭建专属于你的消息推送服务，支持多种消息推送方式，支持 Markdown，仅单可执行文件，开箱即用✨_

</div>

<p align="center">
  <a href="https://raw.githubusercontent.com/songquanpeng/message-pusher/main/LICENSE">
    <img src="https://img.shields.io/github/license/songquanpeng/message-pusher?color=brightgreen" alt="license">
  </a>
  <a href="https://github.com/songquanpeng/message-pusher/releases/latest">
    <img src="https://img.shields.io/github/v/release/songquanpeng/message-pusher?color=brightgreen&include_prereleases" alt="release">
  </a>
  <a href="https://hub.docker.com/repository/docker/justsong/message-pusher">
    <img src="https://img.shields.io/docker/pulls/justsong/message-pusher?color=brightgreen" alt="docker pull">
  </a>
  <a href="https://github.com/songquanpeng/message-pusher/releases/latest">
    <img src="https://img.shields.io/github/downloads/songquanpeng/message-pusher/total?color=brightgreen&include_prereleases" alt="release">
  </a>
  <a href="https://goreportcard.com/report/github.com/songquanpeng/message-pusher">
    <img src="https://goreportcard.com/badge/github.com/songquanpeng/message-pusher" alt="GoReportCard">
  </a>
</p>

<p align="center">
  <a href="https://github.com/songquanpeng/message-pusher/releases">程序下载</a>
  ·
  <a href="#部署">部署教程</a>
  ·
  <a href="#用法">使用教程</a>
  ·
  <a href="https://github.com/songquanpeng/message-pusher/issues">意见反馈</a>
  ·
  <a href="https://push.justsong.cn">在线演示</a>
</p>

> **Fork 说明**：本仓库基于 [songquanpeng/message-pusher](https://github.com/songquanpeng/message-pusher) `master` 分支提交 `fbbfa58`（2025-11-22，`git describe --tags` 为 `v0.4.13-alpha-4-gfbbfa58`，比最新正式版 `v0.4.12` 更新），新增**中国移动新消息（5G 消息 / ClawBot）**原生 Go 推送通道 `cmcc`，协议参考 [x-dr/cmcc-message-gateway](https://github.com/x-dr/cmcc-message-gateway)。属于非官方实现，与中国移动无隶属关系。
>
> - 通道实现：`channel/cmcc.go`，配置方法见[中国移动新消息通道](#中国移动新消息通道)
> - 编译：`.github/workflows/build-cmcc-linux-amd64.yml` 支持 push 到 `main` 自动触发与手动触发，CI 每次从源码构建前端并产出带 SQLite 的 Linux amd64 静态二进制；仓库另附带 `web/build`，便于离线 `go build`
> - 诊断工具：`go build ./cmd/cmcc-probe` 生成独立的链路测试程序
> - `VERSION` = `v0.4.13-alpha-4-gfbbfa58+cmcc`，随仓库提供（上游 CI 用 `git describe --tags` 现场生成，源码仓库原本没有该文件）
> - `docker-compose.cmcc.yml` 基于本地源码构建，官方镜像不含本通道

> **Warning**：原域名（msgpusher.com）将于 2026-11-24 下线，请尽快迁移至 https://push.justsong.cn

> **Warning**：从 `v0.3` 版本升级到 `v0.4` 版本需要手动迁移数据库，具体方法见[迁移数据库](#迁移数据库)。

## 描述
1. **多种消息推送方式**：
   + 邮件消息，
   + 微信测试号，
   + QQ，
   + 企业微信应用号，
   + 企业微信群机器人
   + 飞书自建应用
   + 飞书群机器人，
   + 钉钉群机器人，
   + Bark App,
   + WebSocket 客户端（[官方客户端](https://github.com/songquanpeng/personal-assistant)，[接入文档](./docs/API.md#websocket-客户端)），
   + Telegram 机器人，
   + Discord 群机器人，
   + 中国移动新消息（5G 消息 / ClawBot）：推送到手机「新消息」应用号，
   + 多通道组合：一个通道同时投递到多个推送渠道，
   + **群组消息**：可以将多个推送通道组合成一个群组，然后向群组发送消息，可以实现一次性推送到多个渠道的功能，
   + **自定义消息**：可以自定义消息请求 URL 和请求体格式，实现与其他服务的对接，支持[众多第三方服务](https://iamazing.cn/page/message-pusher-common-custom-templates)。
2. 支持**自定义 Webhook，反向适配各种调用平台**，你可以接入各种已有的系统，而无需修改其代码。
3. 支持在 Web 端编辑 & 管理发送的消息，新消息发送后 Web 端**即时刷新**。
4. 支持**异步**消息发送。
5. 支持用户管理，支持多种用户登录注册方式：
   + 邮箱登录注册以及通过邮箱进行密码重置。
   + [GitHub 开放授权](https://github.com/settings/applications/new)。
   + 微信公众号授权（需要额外部署 [WeChat Server](https://github.com/songquanpeng/wechat-server)）。
6. 支持 Markdown。
7. 支持 Cloudflare Turnstile 用户校验。
8. 支持在线发布公告，设置关于界面以及页脚。
9. API **兼容**其他消息推送服务，例如 [Server 酱](https://sct.ftqq.com/)。

## 用途
1. [整合进自己的博客系统，每当有人登录时发微信消息提醒](https://github.com/songquanpeng/blog/blob/486d63e96ef7906a6c767653a20ec2d3278e9a4a/routes/user.js#L27)。
2. 在进行深度学习模型训练时，在每个 epoch 结束后[将关键数据发送到微信](https://github.com/songquanpeng/pytorch-template/blob/b2ba113659056080d3009b3014a67e977e2851bf/solver/solver.py#L223)以方便及时监控。
3. 在各种脚本运行结束后发消息提醒，例如[监控 GitHub Star 数量的脚本](https://github.com/songquanpeng/scripts/blob/main/star_watcher.py)，又例如[自动健康填报的脚本](https://github.com/songquanpeng/daily-report)，用来通知运行结果。
4. 为[其他系统](https://github.com/songquanpeng/personal-assistant#个人助理应用)提供消息推送功能。

## 部署
### 通过 Docker 部署
部署：`docker run -d --restart always --name message-pusher -p 3000:3000 -e TZ=Asia/Shanghai -v /home/ubuntu/data/message-pusher:/data justsong/message-pusher`

如果无法拉去，请将 `justsong/message-pusher` 替换为 `ghcr.io/songquanpeng/message-pusher`。

更新：`docker run --rm -v /var/run/docker.sock:/var/run/docker.sock containrrr/watchtower -cR`

开放的端口号为 3000，之后用 Nginx 配置域名，反代以及 SSL 证书即可，具体参考[详细部署教程](https://iamazing.cn/page/how-to-deploy-a-website)。

数据将会保存在宿主机的 `/home/ubuntu/data/message-pusher` 目录（只有一个 SQLite 数据库文件），请确保该目录存在且具有写入权限，或者更改为合适的目录。

Nginx 的参考配置：
```
server{
   server_name push.justsong.cn;  # 请根据实际情况修改你的域名
   
   location / {
          client_max_body_size  64m;
          proxy_http_version 1.1;
          proxy_pass http://localhost:3000;  # 请根据实际情况修改你的端口
          proxy_set_header Host $host;
          proxy_set_header X-Forwarded-For $remote_addr;
          proxy_cache_bypass $http_upgrade;
          proxy_set_header Accept-Encoding gzip;
   }
}
```

之后使用 Let's Encrypt 的 certbot 配置 HTTPS：
```bash
# Ubuntu 安装 certbot：
sudo snap install --classic certbot
sudo ln -s /snap/bin/certbot /usr/bin/certbot
# 生成证书 & 修改 Nginx 配置
sudo certbot --nginx
# 根据指示进行操作
# 重启 Nginx
sudo service nginx restart
```

### 手动部署
1. 从 [GitHub Releases](https://github.com/songquanpeng/message-pusher/releases/latest) 下载可执行文件或者从源码编译：
   ```shell
   git clone https://github.com/songquanpeng/message-pusher.git
   cd message-pusher/web
   npm install
   npm run build
   cd ..
   go mod download
   go build -ldflags "-s -w" -o message-pusher
   ````
2. 运行：
   ```shell
   chmod u+x message-pusher
   ./message-pusher --port 3000 --log-dir ./logs
   ```
3. 访问 [http://localhost:3000/](http://localhost:3000/) 并登录。初始账号用户名为 `root`，密码为 `123456`。

如果服务需要长久运行，只是单纯地启动是不够的，[详细部署教程](https://iamazing.cn/page/how-to-deploy-a-website)。

### 注意
如果需要使用 WebSocket 客户端推送功能，则 Nginx 的配置文件中 `proxy_read_timeout` 和 `proxy_send_timeout` 务必设置超过 1 分钟。

推荐设置：
```
proxy_read_timeout 300s;
proxy_send_timeout 300s;   
```

## 配置
系统本身仅需要下载一个可执行文件即可开始使用，无其他依赖。

你可以通过设置环境变量或者命令行参数进行配置。

等到系统启动后，使用 `root` 用户登录系统并做进一步的配置，默认密码为 `123456`。

### 环境变量
1. `REDIS_CONN_STRING`：设置之后将使用 Redis 作为请求频率限制的存储，而非使用内存存储。
    + 例子：`REDIS_CONN_STRING=redis://default:redispw@localhost:49153`
2. `SESSION_SECRET`：设置之后将使用固定的会话密钥，这样系统重新启动后已登录用户的 cookie 将依旧有效。
    + 例子：`SESSION_SECRET=random_string`
3. `SQL_DSN`：设置之后将使用指定数据库而非 SQLite。
    + 例子：`SQL_DSN=root:123456@tcp(localhost:3306)/message-pusher`

注意：使用 Docker 部署时，请使用 `-e key=value` 设置环境变量。 

例子：`docker run -e SESSION_SECRET=random_string ...`

### 命令行参数
1. `--port <port_number>`: 指定服务器监听的端口号，默认为 `3000`。
    + 例子：`--port 3000`
2. `--log-dir <log_dir>`: 指定日志文件夹，如果没有设置，日志将不会被保存。
    + 例子：`--log-dir ./logs`
3. `--version`: 打印系统版本号并退出。


### 进一步的配置
1. 系统设置：
   1. 填写服务器地址。
   2. 配置登录注册选项，如果系统不对外开放，请取消选择`允许新用户注册`。
   3. 配置 SMTP 服务，可以使用 QQ 邮箱的 SMTP 服务。
   4. 其他配置可选，请按照页面上的指示完成配置。
2. 个人设置：
   1. 点击`更新用户信息`更改默认用户名和密码。
   2. 点击`绑定邮箱地址`绑定邮箱以启用邮件消息推送方式。
3. 推送设置：
   1. 设置`默认推送方式`，默认为通过邮件进行推送。
   2. 设置`推送 token`，用以推送 API 调用鉴权，如果不需要留空即可。
   3. 设置其他推送方式，按照页面上的指示即可，完成配置后点击对应的`测试`按钮即可测试配置是否成功。
4. 其他设置：如果系统对外提供服务，本系统也提供了一定的个性化设置功能，你可以设置关于界面和页脚，以及发布公告。

## 用法
1. 消息推送 API URL：`https://<domain>/push/<username>`
   + 将上面的 `<domain>` 以及 `<username>` 替换为真实值，例如：`https://push.mydomain.cn/push/admin`
2. `GET` 请求方式：`https://<domain>/push/<username>?title=<标题>&description=<描述>&content=<Markdown 文本>&channel=<推送方式>&token=<推送 token>`
   1. `title`：选填，受限于具体的消息推送方式，其可能被忽略。
   2. `description`：必填，可以替换为 `desp`。
   3. `content`：选填，受限于具体的消息推送方式，Markdown 语法的支持有所区别。
   4. `channel`：选填，如果不填则系统使用你在后台设置的默认推送通道。注意，此处填的是消息通道的名称，而非类型。可选的推送通道类型有：
      1. `email`：通过发送邮件的方式进行推送（使用 `title` 或 `description` 字段设置邮件主题，使用 `content` 字段设置正文，支持完整的 Markdown 语法）。
      2. `test`：通过微信测试号进行推送（使用 `description` 字段设置模板消息内容，不支持 Markdown）。
      3. `corp_app`：通过企业微信应用号进行推送（仅当使用企业微信 APP 时，如果设置了 `content` 字段，`title` 和 `description` 字段会被忽略；使用微信中的企业微信插件时正常）。
      4. `lark_app`：通过飞书自建应用进行推送。
      5. `corp`：通过企业微信群机器人推送（设置 `content` 字段则将渲染 Markdown 消息，支持 Markdown 的子集；设置 `description` 字段则为普通文本消息）。
      6. `lark`：通过飞书群机器人进行推送（注意事项同上）。
      7. `ding`：通过钉钉群机器人进行推送（注意事项同上）。
      8. `bark`：通过 Bark 进行推送（支持 `title` 和 `description` 字段）。
      9. `client`：通过 WebSocket 客户端进行推送（支持 `title` 和 `description` 字段）。
      10. `telegram`：通过 Telegram 机器人进行推送（`description` 或 `content` 字段二选一，支持 Markdown 的子集）。
      11. `discord`：通过 Discord 群机器人进行推送（注意事项同上）。
      12. `one_api`：通过 OneAPI 协议推送消息到 QQ。
      13. `group`：通过预先配置的消息推送通道群组进行推送。
      14. `custom`：通过预先配置好的自定义推送通道进行推送。
      15. `none`：仅保存到数据库，不做推送。
      16. `cmcc`：通过中国移动「新消息 ClawBot」推送到手机 5G 消息（配置方法见[中国移动新消息通道](#中国移动新消息通道)）。
      17. `multi`：通过一个通道同时投递到多个推送渠道（配置方法见[多通道组合](#多通道组合multi)）。
   5. `token`：如果你在后台设置了推送 token，则此项必填。另外可以通过设置 HTTP `Authorization` 头部设置此项。
      * 注意令牌有两种，一种是全局鉴权令牌，一种是通道维度的令牌，前者可以鉴权任何通道，后者只能鉴权指定通道。
   6. `url`：选填，如果不填则系统自动为消息生成 URL，其内容为消息详情。
   7. `to`：选填，推送给指定用户，如果不填则默认推送给自己，受限于具体的消息推送方式，有些推送方式不支持此项。
      1. `@all`：推送给所有用户。
      2. `user1|user2|user3`：推送给多个用户，用户之间使用 `|` 分隔。
   8. `async`：选填，如果设置为 `true` 则消息推送将在后台异步进行，返回结果包含 `uuid` 字段，可用于后续[获取消息发送状态](./docs/API.md#通过消息 UUID 获取消息发送状态)。
   9. `render_mode`：选填，
      1. 如果设置为 `code`，则消息体会被自动嵌套在代码块中进行渲染；
      2. 如果设置为 `raw`，则不进行 Markdown 解析；
      3. 默认 `markdown`，即进行 Markdown 解析。
3. `POST` 请求方式：字段与上面 `GET` 请求方式保持一致。
   + 如果发送的是 JSON，HTTP Header `Content-Type` 请务必设置为 `application/json`，否则一律按 Form 处理。
   + POST 请求方式下的 `token` 字段也可以通过 URL 查询参数进行设置。


**各种通道的支持程度：**

|      通道类型       | `title` | `description` | `content` | `url` | `to` | Markdown 支持 |
|:---------------:|:-------:|:-------------:|:---------:|:-----:|:----:|:-----------:|
|     `email`     |    ✅    |       ✅       |     ✅     |   ❌   |  ✅️  |     ✅️      |
|     `test`      |    ✅    |       ✅       |     ✅     |  ✅️   |  ✅️  |      ✅      |
|   `corp_app`    |    ✅    |       ✅       |     ✅     |  ✅️   |  ✅   |      ✅      |
|     `corp`      |    ❌    |       ✅       |     ✅     |  ✅️   |  ✅️  |      ✅      |
|     `lark`      |    ❌    |       ✅       |     ✅     |   ❌   |  ✅   |      ✅      |
|   `lark_app`    |    ❌    |       ✅       |     ✅     |  ❌️   |  ✅   |      ✅      |
|     `ding`      |    ✅    |       ✅       |     ✅     |  ✅️   |  ✅   |      ✅      |
|     `bark`      |    ✅    |       ✅       |     ✅     |  ✅️   |  ❌   |      ✅      |
|    `client`     |    ✅    |       ✅       |     ❌     |   ❌   |  ❌   |      ❌      |
|   `telegram`    |    ❌    |       ❌       |     ✅     |   ❌   |  ✅   |      ✅      |
|    `discord`    |    ❌    |       ❌       |     ✅     |   ❌   |  ✅   |      ❌      |
|     `cmcc`      |    ✅    |       ✅       |     ✅     |  ✅️   |  ✅   |      ❌      |
|     `multi`     |    ✅    |       ✅       |     ✅     |  ✅️   |  ✅   |  取决于子渠道  |

注意：
1. 对于大部分通道，`description` 字段和 `content` 是不能同时存在的，如果你只需要文字消息，请使用 `description` 字段，如果你需要发送 Markdown 消息，请使用 `content` 字段。
2. 部分通道的 Markdown 支持实际上是通过 URL 跳转到本系统所渲染的消息详情实现的，其他通道的 Markdown 支持受限于具体的通道，支持的语法并不统一。

### 中国移动新消息通道

该通道把消息推送到手机「新消息」应用号，协议参考 [cmcc-message-gateway](https://github.com/x-dr/cmcc-message-gateway)，属于**非官方实现**，与中国移动无隶属关系。使用前请确认你的号码与终端支持「新消息 ClawBot」，并遵守相关服务条款。

**获取 API Key**：用手机打开[中国移动官方文档](https://f.10086.cn/s/s/msgNotice/?infoId=oGaKNuZqtg6KbLn9OyRceA%3d%3d)，进入「新消息 ClawBot」应用号 → 绑定/解绑 → 立即授权，按提示完成认证后取得专属 API Key。请妥善保管，不要写入日志或提交到仓库。

**通道配置**（后台 → 新建消息通道 → 类型选「中国移动新消息」）：

| 表单项 | 对应字段 | 说明 |
| --- | --- | --- |
| API Key | `secret` | 必填，上面取得的 `CMCC_API_KEY` |
| 默认接收人 | `account_id` | 选填；不填时自动使用最近一条入站消息的发送方 |
| 协议版本 | `app_id` | 选填，默认 `2.0` |
| WebSocket 服务地址 | `url` | 选填，默认 `wss://5gvas01.cmicmaap.com/gtw-ai/openclaw/ws/msg` |
| 入站消息 Webhook | `other` | 选填，收到入站消息时把 JSON POST 到该地址 |

接收人优先级：推送请求的 `to` 参数 > 通道默认接收人 > 最近一条入站消息的发送方。三者都没有时，先用手机向该应用号发送一条任意消息（或推送时显式传 `to`）即可。

推送时 `title` 与 `description`/`content` 会合并为一段纯文本，Markdown 标记会被降级为纯文本；`render_mode=raw` 时保留原文。`url` 字段在本通道被解释为**富媒体直链**：当它是可识别的图片、音视频或文件直链时发送对应的媒体消息，否则忽略。发送成功只代表消息已写入 WebSocket，不代表手机端已送达或已读。

本通道由 Go 进程直接维护 WebSocket 长连接（含认证、心跳与断线重连），**无需额外部署 Cloudflare Worker**。

### 多通道组合（multi）

一个通道对应多个推送渠道，仍然只有一个鉴权令牌：调用一次 `/push` 并指定该通道，即可把同一条消息投递到所有子渠道。与 `group` 群组通道的区别是：`multi` 不需要预先创建子通道，所有配置都在本通道内完成。

在后台新建通道时，类型选「多通道组合」，然后在「勾选渠道」下拉里**一次选中多个渠道**（例如「企业微信群机器人」+「中国移动新消息」），点击「添加选中的渠道」。每选中一个渠道就会生成一张配置卡片，直接在该卡片里填写对应参数即可，无需手写 JSON。同一种渠道可以重复添加多次（例如两个不同的企业微信群）。卡片支持上移/下移调整发送顺序、删除，也能在卡片里直接改渠道类型。

已配置好的通道再次编辑时，会自动还原成卡片列表；勾选「高级：直接编辑 JSON 配置」可查看或修改底层 JSON，其格式为一个数组，每项的 `type` 与单通道类型一致，其余字段含义也与单通道相同：

```json
[
  {
    "type": "corp",
    "url": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=你的key"
  },
  {
    "type": "cmcc",
    "secret": "你的 CMCC API Key",
    "account_id": "接收人 ID"
  }
]
```

上面的例子就是「企业微信群机器人 + 移动短信」同时推送。可视化卡片与这段 JSON 完全等价。

说明：
1. 渠道按卡片顺序依次发送；某一个失败不影响其余渠道继续发，全部发完后汇总返回失败信息。
2. 渠道类型不能是 `multi` 或 `group`（下拉里已排除），避免递归；`email` 由系统自动创建、`none` 不推送，也不在可选列表内。
3. 若推送请求带了 `to`，所有渠道都会收到同一个 `to`；需要各自指定目标时请写在各卡片的「默认接收人 / 默认推送目标」中，且调用时不要传 `to`。
4. 只要有一个渠道失败，该条消息会被标记为失败，但其他渠道实际已投递。

**示例：**

<details>
<summary><strong>Bash 示例 </strong></summary>
<div>

```shell
#!/bin/bash

MESSAGE_PUSHER_SERVER="https://push.justsong.cn"
MESSAGE_PUSHER_USERNAME="test"
MESSAGE_PUSHER_TOKEN="666"

function send_message {
  # POST Form
  curl -s -X POST "$MESSAGE_PUSHER_SERVER/push/$MESSAGE_PUSHER_USERNAME" \
    -d "title=$1&description=$2&content=$3&token=$MESSAGE_PUSHER_TOKEN" \
    >/dev/null
}

function send_message_with_json {
  # POST JSON
  curl -s -X POST "$MESSAGE_PUSHER_SERVER/push/$MESSAGE_PUSHER_USERNAME" \
    -H 'Content-Type: application/json' \
    -d '{"title":"'"$1"'","desp":"'"$2"'", "content":"'"$3"'", "token":"'"$MESSAGE_PUSHER_TOKEN"'"}' \
    >/dev/null
}

send_message 'title' 'description' 'content'
```

另一个版本：
```shell
MESSAGE_PUSHER_SERVER="https://push.justsong.cn"
MESSAGE_PUSHER_USERNAME="test"
MESSAGE_PUSHER_TOKEN="666"
MESSAGE_PUSHER_CHANNEL="lark"

sendmsg() {
    if [ -t 0 ]; then
        local param="$*"
    else
        local param=$(</dev/stdin)
    fi
    curl -s -o /dev/null --get --data-urlencode "content=${param}" "$MESSAGE_PUSHER_SERVER/push/$MESSAGE_PUSHER_USERNAME?channel=$MESSAGE_PUSHER_CHANNEL&token=$MESSAGE_PUSHER_TOKEN"
}
```

之后便可以进行这样的操作：
```shell
uname -ra | sendmsg
```

</div>
</details>

<details>
<summary><strong>Python 示例 </strong></summary>
<div>

```python
import requests

SERVER = "https://push.justsong.cn"
USERNAME = "test"
TOKEN = "666"


def send_message(title, description, content):
    # GET 方式
    # res = requests.get(f"{SERVER}/push/{USERNAME}?title={title}"
    #                    f"&description={description}&content={content}&token={TOKEN}")

    # POST 方式
    res = requests.post(f"{SERVER}/push/{USERNAME}", json={
        "title": title,
        "description": description,
        "content": content,
        "token": TOKEN
    })
    res = res.json()
    if res["success"]:
        return None
    else:
        return res["message"]


error = send_message("标题", "描述", "**Markdown 内容**")
if error:
    print(error)
```

</div>
</details>

<details>
<summary><strong>Go 示例 </strong></summary>
<div>

```go
package main

import (
   "bytes"
   "encoding/json"
   "errors"
   "fmt"
   "net/http"
   "net/url"
)

var serverAddress = "https://push.justsong.cn"
var username = "test"
var token = "666"

type request struct {
   Title       string `json:"title"`
   Description string `json:"description"`
   Content     string `json:"content"`
   URL         string `json:"url"`
   Channel     string `json:"channel"`
   Token       string `json:"token"`
}

type response struct {
   Success bool   `json:"success"`
   Message string `json:"message"`
}

func SendMessage(title string, description string, content string) error {
   req := request{
      Title:       title,
      Description: description,
      Content:     content,
      Token:       token,
   }
   data, err := json.Marshal(req)
   if err != nil {
      return err
   }
   resp, err := http.Post(fmt.Sprintf("%s/push/%s", serverAddress, username),
      "application/json", bytes.NewBuffer(data))
   if err != nil {
      return err
   }
   var res response
   err = json.NewDecoder(resp.Body).Decode(&res)
   if err != nil {
      return err
   }
   if !res.Success {
      return errors.New(res.Message)
   }
   return nil
}

func SendMessageWithForm(title string, description string, content string) error {
   resp, err := http.PostForm(fmt.Sprintf("%s/push/%s", serverAddress, username),
      url.Values{"title": {title}, "description": {description}, "content": {content}, "token": {token}})
   if err != nil {
      return err
   }
   var res response
   err = json.NewDecoder(resp.Body).Decode(&res)
   if err != nil {
      return err
   }
   if !res.Success {
      return errors.New(res.Message)
   }
   return nil
}

func main() {
   //err := SendMessage("标题", "描述", "**Markdown 内容**")
   err := SendMessageWithForm("标题", "描述", "**Markdown 内容**")
   if err != nil {
      fmt.Println("推送失败：" + err.Error())
   } else {
      fmt.Println("推送成功！")
   }
}
```

</div>
</details>

<details>
<summary><strong>C# 示例 </strong></summary>
<div>

```csharp
using Newtonsoft.Json;
using RestSharp;

namespace Demo
{
    public class Program
    {
        public static void Main(string[] args)
        {
            //推送消息
            var sendMsg = MessagePusherTool.SendMessage("标题", "描述", "**Markdown 内容**");
            if(sendMsg.Success)
            {
                Console.WriteLine($"推送成功！");
            }
            else
            {
                Console.WriteLine($"推送失败：{sendMsg.Message}");
            }
        }
    }

    /// <summary>
    /// 消息推送工具
    /// 
    /// <para>开源地址：https://github.com/songquanpeng/message-pusher</para>
    /// <para>支持：Framework、Net3.1、Net5、Net6</para>
    /// <para>引用包：</para>
    /// <para>dotnet add package Newtonsoft.Json -v 13.0.2</para>
    /// <para>dotnet add package RestSharp -v 108.0.3</para>
    /// </summary>
    public class MessagePusherTool
    {
        /// <summary>
        /// ServerAddress
        /// </summary>
        public const string ServerAddress = "https://push.justsong.cn";

        /// <summary>
        /// UserName
        /// </summary>
        public const string UserName = "test";

        /// <summary>
        /// Token
        /// </summary>
        public const string Token = "666";

        /// <summary>
        /// SendMessage
        /// </summary>
        /// <param name="title">title</param>
        /// <param name="description">description</param>
        /// <param name="content">content</param>
        public static Response SendMessage(string title, string description, string content)
        {
            var requestData = new Request()
            {
                Title = title,
                Description = description,
                Content = content,
                Token = Token,
            };
            var url = $"{ServerAddress}";
            var client = new RestClient(url);
            var request = new RestRequest($"push/{UserName}", Method.Post);
            request.AddJsonBody(requestData);
            var response = client.Execute(request);
            var responseData = response.Content;
            var responseJson = JsonConvert.DeserializeObject<Response>(responseData);
            return responseJson;
        }

        /// <summary>
        /// Request
        /// </summary>
        public class Request
        {
            /// <summary>
            /// Title
            /// </summary>
            [JsonProperty(PropertyName = "title")]
            public string Title { get; set; }

            /// <summary>
            /// Description
            /// </summary>
            [JsonProperty(PropertyName = "description")]
            public string Description { get; set; }

            /// <summary>
            /// Content
            /// </summary>
            [JsonProperty(PropertyName = "content")]
            public string Content { get; set; }

            /// <summary>
            /// URL
            /// </summary>
            [JsonProperty(PropertyName = "url")]
            public string URL { get; set; }

            /// <summary>
            /// Channel
            /// </summary>
            [JsonProperty(PropertyName = "channel")]
            public string Channel { get; set; }

            /// <summary>
            /// Token
            /// </summary>
            [JsonProperty(PropertyName = "token")]
            public string Token { get; set; }
        }

        /// <summary>
        /// Response
        /// </summary>
        public class Response
        {
            /// <summary>
            /// Success
            /// </summary>
            [JsonProperty(PropertyName = "success")]
            public bool Success { get; set; }

            /// <summary>
            /// Message
            /// </summary>
            [JsonProperty(PropertyName = "message")]
            public string Message { get; set; }
        }
    }
}
```

</div>
</details>

<details>
<summary><strong>Node.js 示例 </strong></summary>
<div>

```javascript
const axios = require('axios');
const querystring = require('querystring');

const MESSAGE_PUSHER_SERVER = 'https://push.justsong.cn'
const MESSAGE_PUSHER_USERNAME = 'test'
const MESSAGE_PUSHER_TOKEN = '666'

async function send_message(title, description, content) {
  try {
    const postData = querystring.stringify({
      title: title,
      desp: description,
      content: content,
      token: MESSAGE_PUSHER_TOKEN,
    })

    const response = await axios.post(`${MESSAGE_PUSHER_SERVER}/push/${MESSAGE_PUSHER_USERNAME}`, postData, {
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
    })
    if (response.data.success) {
      return response.data
    }
  } catch (error) {
    if (error.response) {
      return error.response.data
    } else {
      throw error
    }

  }
}

send_message('标题', '描述', '**Markdown 内容**')
  .then((response) => {
    if (response.success) {
      console.log('推送成功:', response)
    } else {
      console.log('推送失败:', response)
    }
  }, (error) => {
    console.log(error.message);
  })

```
</div>
</details>

欢迎 PR 添加更多语言的示例。


## 迁移数据库
此处均以 SQLite 为例，其他数据库请自行修改。我已经让 ChatGPT 翻译成对应的 SQL 版本，见 `bin` 文件夹，供参考。

### 从 `v0.3` 迁移到 `v0.4`
1. 首先备份你的数据库文件。
2. 下载最新的 `v0.4` 版本，启动程序，程序会自动进行数据库表结构的迁移。
3. 终止程序。
4. 之后执行脚本：`./bin/migrate_v3_to_v4.py`，进行数据的迁移。
5. 重新启动程序即可。

注意，执行前请确保数据库中 `users` 表中字段的顺序和脚本中的一致，否则会出现数据错乱的情况。

## 其他
1. `v0.3` 之前的版本基于 Node.js，你可以切换到 [`nodejs`](https://github.com/songquanpeng/message-pusher/tree/nodejs) 分支查看，该版本不再有功能性更新。
2. `v0.3` 以及后续版本基于 Gin Template [`v0.2.1`](https://github.com/songquanpeng/gin-template) 版本开发。
3. 如果想要自行编译，请首先[编译前端](./web/README.md)，之后再编译后端，否则会遇到 `pattern web/build: no matching files found` 问题。
