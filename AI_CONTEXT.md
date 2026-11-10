# message-pusher 定制版 · 项目交接说明

> 供 AI 助手接手本项目时阅读。最后更新：2026-11-10，对应版本 `v0.5.1`。

## 一、项目是什么

自托管的统一消息推送服务。基于开源项目 [songquanpeng/message-pusher](https://github.com/songquanpeng/message-pusher) 的定制分支：单个 Go 可执行文件内嵌 React 前端，默认 SQLite，提供 `/push/<username>` 推送 API 与 Web 管理后台，可把消息投递到邮件、企业微信、飞书、钉钉、Bark、Telegram、Discord、WebSocket 客户端、中国移动 5G 消息等渠道。

| 项 | 值 |
| --- | --- |
| 公开仓库 | https://github.com/liuchenyao114/123 |
| 上游基准 | `master@fbbfa58`（2025-11-22，`v0.4.13-alpha-4-gfbbfa58`，比最新正式版 `v0.4.12` 新） |
| 当前正式版本 | `v0.5.1` |
| 下载地址 | https://github.com/liuchenyao114/123/releases/latest |
| 技术栈 | Go 1.22 + Gin + GORM（SQLite / MySQL）+ gorilla/websocket；React + semantic-ui-react；前端产物用 `go:embed` 内嵌 |
| 许可证 | MIT（沿用上游） |

属于**非官方实现**，与中国移动无隶属关系。

## 二、相对上游的定制内容

| 变化 | 说明 |
| --- | --- |
| 新增 `cmcc` 通道 | Go 原生 WebSocket 对接中国移动「新消息 ClawBot / 5G 消息」，含握手 `X-API-Key`、`auth`/`auth_ok`、15 秒心跳、断线重连、入站消息记录接收人、Webhook 转发、Markdown 降级、富媒体直链。协议参考 [x-dr/cmcc-message-gateway](https://github.com/x-dr/cmcc-message-gateway)，**无需额外部署 Cloudflare Worker 或网关** |
| Webhook 可选多个通道 | 「接口配置」页的通道下拉可多选；一次请求把同一条消息分别投递到所有选中通道，某个通道失败不影响其余通道；`/push` 也支持 `channel=a|b` |
| 移除 `tencent_alarm` | 腾讯云自定义告警 / 短信通道的代码、界面、文档全部删除 |
| 修复鉴权绕过 | 旧逻辑在「设了全局令牌、通道未设令牌」时不带 token 也能推送；现改为必须匹配全局令牌或该通道令牌 |
| 构建与发布自动化 | push 到 `main` 自动重建前端 + 跑单元测试 + 编译 linux/amd64；打 `v*` tag 或手动触发自动发布 4 平台正式 Release |

**重要**：`multi` 组合通道曾短暂加入，现已**完全回滚**。用户明确要求「新建消息通道」页保持上游原始逻辑（一个通道选一种类型），多选能力只在 Webhook 接口页与 `channel=a|b` 中提供。请勿重新引入。

## 三、关键文件

| 路径 | 作用 |
| --- | --- |
| `channel/cmcc.go` | cmcc 通道实现（长连接、认证、心跳、重连、入站、富媒体） |
| `controller/message.go` | `splitChannelNames` / `processMultiChannelMessage` / `authMessage`，多通道投递与鉴权 |
| `controller/webhook.go` | `TriggerWebhook`，服务内部触发，**不校验 token** |
| `model/webhook.go` | `channel` 列 `varchar(1024)`，多个通道名用 `|` 分隔 |
| `web/src/pages/Webhook/EditWebhook.js` | Webhook 编辑页的通道多选下拉 |
| `web/src/pages/Channel/EditChannel.js` | 新建/编辑通道页，保持上游原始逻辑 + `cmcc` 表单 |
| `.github/workflows/build-cmcc-linux-amd64.yml` | push main 自动构建 + 单元测试 |
| `.github/workflows/release.yml` | 4 平台构建并自动发布正式 Release |
| `cmd/cmcc-probe/main.go` | 独立的 cmcc 链路诊断工具 |
| `VERSION` / `docker-compose.cmcc.yml` | 版本标记 / 基于本地源码的编排文件 |
| `controller/message_test.go` | 鉴权与通道名拆分的表驱动测试 |

## 四、构建与发布

全部在 GitHub Actions 上完成，本地不需要 Go / Node 环境。

| 触发 | 工作流 | 产物 |
| --- | --- | --- |
| push 到 `main` | `build-cmcc-linux-amd64.yml` | 从源码重建前端 → 单元测试 → 编译带 SQLite 的静态 linux/amd64 二进制，上传 Actions Artifact |
| 打 `v*` tag 或手动 dispatch | `release.yml` | 4 个平台二进制，自动创建正式 Release（非 draft） |

产物命名（**不带 cmcc 后缀**）：`message-pusher-linux-amd64`、`message-pusher-linux-arm64`、`message-pusher-windows-amd64.exe`、`message-pusher-darwin-arm64`。

CI 每次都从源码构建前端；仓库内附带的 `web/build` 仅供离线 `go:embed` 编译，可能落后于源码。单元测试用 `go test -c` 编译后不带参数运行，以避开 `common/init.go` 中 `flag.Parse()` 与 `go test` 参数的冲突。

## 五、运行与调用

```bash
chmod +x message-pusher-linux-amd64
./message-pusher-linux-amd64 --port 3000
```

浏览器打开 `http://<服务器IP>:3000`，初始账号 `root` / `123456`，**登录后立即改密**；对外开放时关闭注册、设置推送令牌、启用 HTTPS。

```bash
# GET
curl "http://127.0.0.1:3000/push/root?title=告警&description=磁盘90%&channel=corp_bot|cmcc_phone&token=<全局令牌>"
# POST JSON
curl -X POST http://127.0.0.1:3000/push/root -H 'Content-Type: application/json' \
  -d '{"title":"告警","content":"**磁盘** 90%","channel":"corp_bot","token":"<全局令牌>"}'
```

常用参数：`title`、`description`（兼容 `desp`/`short`）、`content`、`url`、`channel`、`token`、`to`、`async`、`render_mode`。`Authorization: Bearer <token>` 亦可。异步推送返回 `uuid` 可查状态；多通道返回 `results: [{channel, uuid, error}]`。

**鉴权判定**（仅针对直接调用 `/push`；Webhook 触发不校验）：

| 设置情况 | 是否要带 token |
| --- | --- |
| 全局令牌、通道令牌都没设 | 不用带 |
| 设了全局令牌 | 必须带全局令牌（或该通道自己的令牌） |
| 通道设了独立令牌 | 必须带那个通道的令牌（或全局令牌） |
| 多通道 `channel=a|b` | 建议带全局令牌，每个通道独立校验 |

全局令牌在「推送设置」页设置；通道令牌在「新建/编辑通道」页的「鉴权令牌」里设置。

## 六、升级到本版本的注意事项

1. 前端内嵌在二进制中，**必须替换整个可执行文件并重启进程**，只改配置不会更新界面。
2. 鉴权收紧：升级前确认所有 `/push` 调用方都带了全局令牌，否则会返回 401。Webhook 触发不受影响。
3. 老数据库执行 `DELETE FROM channels WHERE type IN ('tencent_alarm','multi');` 清理失效通道记录。
4. `webhooks.channel` 扩为 `varchar(1024)`：MySQL 由 GORM AutoMigrate 自动改列（需账号有 ALTER 权限），SQLite 无需处理。数据不会丢失。
5. Docker 部署必须先设置 `SESSION_SECRET`（`export SESSION_SECRET=$(openssl rand -hex 32)`），否则 compose 拒绝启动；更换密钥后需重新登录。
6. cmcc 属非官方实现，需号码/终端支持 ClawBot 并自行申请 API Key；写入 WebSocket 成功不代表手机已送达；建议单实例部署。

## 七、给接手 AI 的约定

- 用户的工作流偏好：**只负责写好代码 → 同步到 GitHub → 云端构建完成后提醒**，不要在本地构建、下载产物或反复打包。
- 不要改动「新建消息通道」页的原有逻辑，不要重新引入 `multi` 通道。
- 多选能力只在 Webhook 接口页与 `/push` 的 `channel=a|b` 两处。
- 仓库为公开仓库，任何 AI 可直接通过 `raw.githubusercontent.com` 读取本文件与全部源码。
- 本文件及仓库不包含任何访问凭证；API Key、推送令牌均存在运行时的数据库中。
