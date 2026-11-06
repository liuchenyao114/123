package channel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"message-pusher/common"
	"message-pusher/model"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// CMCC 5G 消息（新消息 ClawBot）通道。
// 协议参考：https://github.com/x-dr/cmcc-message-gateway
// 与 Cloudflare Worker 版不同，这里在 Go 进程内直接维护 WebSocket 长连接，
// 因此不需要额外部署网关服务。
const (
	cmccDefaultServerURL = "wss://5gvas01.cmicmaap.com/gtw-ai/openclaw/ws/msg"
	cmccDefaultVersion   = "2.0"
	cmccAuthTimeout      = 10 * time.Second
	cmccDialTimeout      = 15 * time.Second
	cmccHeartbeatPeriod  = 15 * time.Second
	cmccPongTimeout      = 45 * time.Second
	cmccWriteTimeout     = 10 * time.Second
	cmccWebhookTimeout   = 10 * time.Second
)

var (
	cmccConnectionsMutex sync.Mutex
	cmccConnections      = make(map[string]*cmccConnection)
	cmccDialMutex        sync.Mutex // 仅串行化建连，避免同一 API Key 建立多条连接
	cmccMessageSeq       uint64
)

type cmccFrame struct {
	Type      string `json:"type"`
	APIKey    string `json:"apiKey,omitempty"`
	Version   string `json:"version,omitempty"`
	To        string `json:"to,omitempty"`
	Content   string `json:"content,omitempty"`
	MessageID string `json:"messageId,omitempty"`
	MediaType string `json:"mediaType,omitempty"`
	MediaURL  string `json:"mediaUrl,omitempty"`
	From      string `json:"from,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Message   string `json:"message,omitempty"`
	Text      string `json:"text,omitempty"`
	Timestamp int64  `json:"timestamp,omitempty"`
}

type cmccConfig struct {
	key              string // 同一 API Key + 服务地址复用一条连接
	apiKey           string
	serverURL        string
	version          string
	defaultRecipient string
	webhook          string
}

type cmccConnection struct {
	key       string
	apiKey    string
	serverURL string
	version   string

	conn        *websocket.Conn
	writeMutex  sync.Mutex // gorilla/websocket 只允许一个并发 writer
	authResult  chan error
	stop        chan struct{}
	stopOnce    sync.Once
	lastPong    int64 // atomic，UnixMilli
	webhook     string
	runtimeLock sync.RWMutex

	recipient     string // 最近一次入站消息的发送方，自动作为默认接收人
	recipientLock sync.RWMutex
}

func parseCmccConfig(channel_ *model.Channel) cmccConfig {
	apiKey := strings.TrimSpace(channel_.Secret)
	serverURL := strings.TrimSpace(channel_.URL)
	if serverURL == "" {
		serverURL = cmccDefaultServerURL
	}
	version := strings.TrimSpace(channel_.AppId)
	if version == "" {
		version = cmccDefaultVersion
	}
	return cmccConfig{
		key:              apiKey + "|" + serverURL,
		apiKey:           apiKey,
		serverURL:        serverURL,
		version:          version,
		defaultRecipient: strings.TrimSpace(channel_.AccountId),
		webhook:          strings.TrimSpace(channel_.Other),
	}
}

func getCmccConnection(cfg cmccConfig) (*cmccConnection, error) {
	cmccConnectionsMutex.Lock()
	c, ok := cmccConnections[cfg.key]
	cmccConnectionsMutex.Unlock()
	if ok && !c.isClosed() {
		c.setWebhook(cfg.webhook)
		return c, nil
	}

	cmccDialMutex.Lock()
	defer cmccDialMutex.Unlock()
	// 双重检查，避免并发建连
	cmccConnectionsMutex.Lock()
	c, ok = cmccConnections[cfg.key]
	cmccConnectionsMutex.Unlock()
	if ok && !c.isClosed() {
		c.setWebhook(cfg.webhook)
		return c, nil
	}

	conn, err := dialCmcc(cfg)
	if err != nil {
		return nil, err
	}
	cmccConnectionsMutex.Lock()
	// 建连期间可能已有其他 goroutine 完成建连，保留先建立的那条
	if existing, exists := cmccConnections[cfg.key]; exists && !existing.isClosed() {
		cmccConnectionsMutex.Unlock()
		conn.close()
		return existing, nil
	}
	cmccConnections[cfg.key] = conn
	cmccConnectionsMutex.Unlock()
	return conn, nil
}

func dialCmcc(cfg cmccConfig) (*cmccConnection, error) {
	header := http.Header{}
	header.Set("X-API-Key", cfg.apiKey)
	dialer := websocket.Dialer{HandshakeTimeout: cmccDialTimeout}
	conn, resp, err := dialer.Dial(cfg.serverURL, header)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("CMCC WebSocket 握手失败（HTTP %d）：%w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("CMCC WebSocket 握手失败：%w", err)
	}

	c := &cmccConnection{
		key:        cfg.key,
		apiKey:     cfg.apiKey,
		serverURL:  cfg.serverURL,
		version:    cfg.version,
		conn:       conn,
		authResult: make(chan error, 1),
		stop:       make(chan struct{}),
		webhook:    cfg.webhook,
	}
	atomic.StoreInt64(&c.lastPong, time.Now().UnixMilli())

	// 握手完成后先启动读循环，再发送 auth，避免错过 auth_ok
	go c.readLoop()
	if err = c.write(cmccFrame{Type: "auth", APIKey: cfg.apiKey, Version: cfg.version}); err != nil {
		c.close()
		return nil, fmt.Errorf("CMCC 认证帧发送失败：%w", err)
	}

	select {
	case err = <-c.authResult:
		if err != nil {
			c.close()
			return nil, err
		}
	case <-time.After(cmccAuthTimeout):
		c.close()
		return nil, errors.New("等待 CMCC auth_ok 超时，请检查 API Key 与服务地址")
	}

	go c.heartbeatLoop()
	common.SysLog("CMCC 5G 消息通道已连接并认证成功")
	return c, nil
}

func (c *cmccConnection) readLoop() {
	defer c.close()
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			select {
			case <-c.stop:
			default:
				common.SysError("CMCC WebSocket 读取失败，连接将重建：" + err.Error())
			}
			return
		}
		var frame cmccFrame
		if err = json.Unmarshal(data, &frame); err != nil {
			continue
		}
		switch frame.Type {
		case "auth_ok":
			select {
			case c.authResult <- nil:
			default:
			}
		case "auth_failed":
			message := frame.Message
			if message == "" {
				message = "CMCC 认证失败，请检查 API Key"
			}
			select {
			case c.authResult <- errors.New(message):
			default:
			}
			return
		case "pong":
			atomic.StoreInt64(&c.lastPong, time.Now().UnixMilli())
		case "ping":
			if err = c.write(cmccFrame{Type: "pong"}); err != nil {
				return
			}
		case "text_message", "media_message", "message":
			c.handleInbound(frame)
		}
	}
}

func (c *cmccConnection) heartbeatLoop() {
	ticker := time.NewTicker(cmccHeartbeatPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-c.stop:
			return
		case <-ticker.C:
			last := atomic.LoadInt64(&c.lastPong)
			if last > 0 && time.Since(time.UnixMilli(last)) > cmccPongTimeout {
				common.SysError("CMCC pong 超时，连接将重建")
				c.close()
				return
			}
			if err := c.write(cmccFrame{Type: "ping"}); err != nil {
				c.close()
				return
			}
		}
	}
}

func (c *cmccConnection) handleInbound(frame cmccFrame) {
	sender := strings.TrimSpace(frame.From)
	if sender == "" {
		sender = strings.TrimSpace(frame.Phone)
	}
	// CMCC 的状态回包也叫 message，但没有发送方，不能覆盖接收人
	if sender == "" {
		return
	}
	c.setRecipient(sender)

	content := frame.Content
	if content == "" {
		content = frame.Text
	}
	if content == "" {
		content = frame.Message
	}
	c.runtimeLock.RLock()
	webhook := c.webhook
	c.runtimeLock.RUnlock()
	if webhook == "" {
		return
	}
	payload := map[string]interface{}{
		"from":      sender,
		"type":      frame.Type,
		"content":   content,
		"mediaType": frame.MediaType,
		"mediaUrl":  frame.MediaURL,
		"timestamp": time.Now().UnixMilli(),
	}
	go c.deliverWebhook(webhook, payload)
}

func (c *cmccConnection) deliverWebhook(webhook string, payload map[string]interface{}) {
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	client := &http.Client{Timeout: cmccWebhookTimeout}
	resp, err := client.Post(webhook, "application/json", bytes.NewReader(body))
	if err != nil {
		common.SysError("CMCC 入站 Webhook 转发失败：" + err.Error())
		return
	}
	_ = resp.Body.Close()
}

func (c *cmccConnection) write(frame cmccFrame) error {
	c.writeMutex.Lock()
	defer c.writeMutex.Unlock()
	select {
	case <-c.stop:
		return errors.New("CMCC 连接已关闭")
	default:
	}
	_ = c.conn.SetWriteDeadline(time.Now().Add(cmccWriteTimeout))
	return c.conn.WriteJSON(frame)
}

func (c *cmccConnection) isClosed() bool {
	select {
	case <-c.stop:
		return true
	default:
		return false
	}
}

func (c *cmccConnection) close() {
	c.stopOnce.Do(func() {
		close(c.stop)
		c.writeMutex.Lock()
		_ = c.conn.Close()
		c.writeMutex.Unlock()
		cmccConnectionsMutex.Lock()
		if current, ok := cmccConnections[c.key]; ok && current == c {
			delete(cmccConnections, c.key)
		}
		cmccConnectionsMutex.Unlock()
	})
}

// CloseCmccConnectionsByAPIKey 关闭指定 API Key 的所有长连接。
// 通道被删除，或 API Key / 服务地址被修改时调用，避免旧连接继续占用资源。
func CloseCmccConnectionsByAPIKey(apiKey string) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return
	}
	cmccConnectionsMutex.Lock()
	var targets []*cmccConnection
	for _, c := range cmccConnections {
		if c.apiKey == apiKey {
			targets = append(targets, c)
		}
	}
	cmccConnectionsMutex.Unlock()
	for _, c := range targets {
		c.close()
	}
}

func (c *cmccConnection) setWebhook(webhook string) {
	c.runtimeLock.Lock()
	c.webhook = webhook
	c.runtimeLock.Unlock()
}

func (c *cmccConnection) setRecipient(recipient string) {
	c.recipientLock.Lock()
	c.recipient = recipient
	c.recipientLock.Unlock()
}

func (c *cmccConnection) getRecipient() string {
	c.recipientLock.RLock()
	defer c.recipientLock.RUnlock()
	return c.recipient
}

func nextCmccMessageID() string {
	seq := atomic.AddUint64(&cmccMessageSeq, 1)
	return fmt.Sprintf("msg_%d_%d", time.Now().UnixMilli(), seq)
}

// SendCmccMessage 通过中国移动「新消息 ClawBot」把消息推送到手机。
// 接收人优先级：推送请求的 to 参数 > 通道配置的默认接收人 > 最近一条入站消息的发送方。
func SendCmccMessage(message *model.Message, user *model.User, channel_ *model.Channel) error {
	cfg := parseCmccConfig(channel_)
	if cfg.apiKey == "" {
		return errors.New("CMCC 通道缺少 API Key，请在通道的「API Key」中填写")
	}
	if !strings.HasPrefix(cfg.serverURL, "ws://") && !strings.HasPrefix(cfg.serverURL, "wss://") {
		return errors.New("CMCC 服务地址必须以 ws:// 或 wss:// 开头")
	}

	content := buildCmccContent(message)
	if content == "" {
		return errors.New("CMCC 消息内容不能为空")
	}

	conn, err := getCmccConnection(cfg)
	if err != nil {
		return err
	}

	to := strings.TrimSpace(message.To)
	if to == "" {
		to = cfg.defaultRecipient
	}
	if to == "" {
		to = conn.getRecipient()
	}
	if to == "" {
		return errors.New("尚未确定 CMCC 接收人：请在通道中填写默认接收人，或先用手机向 5G 消息应用号发送一条消息，或在推送时指定 to 参数")
	}

	frame := cmccFrame{
		Type:      "send",
		APIKey:    cfg.apiKey,
		To:        to,
		Content:   content,
		MessageID: nextCmccMessageID(),
	}
	// CMCC 通道把消息 url 字段解释为富媒体直链：只有可识别的媒体文件才走媒体帧
	if mediaType, ok := inferCmccMediaType(message.URL); ok {
		frame.MediaType = mediaType
		frame.MediaURL = strings.TrimSpace(message.URL)
	}

	if err = conn.write(frame); err != nil {
		// 连接可能在两次心跳之间断开，重连后重试一次
		common.SysError("CMCC 发送失败，尝试重连：" + err.Error())
		oldRecipient := conn.getRecipient()
		conn.close()
		retryConn, retryErr := getCmccConnection(cfg)
		if retryErr != nil {
			return retryErr
		}
		if oldRecipient != "" {
			retryConn.setRecipient(oldRecipient)
		}
		return retryConn.write(frame)
	}
	return nil
}

func buildCmccContent(message *model.Message) string {
	body := message.Content
	if body == "" {
		body = message.Description
	}
	if message.RenderMode != "raw" {
		body = cmccPlainText(body)
	}
	title := strings.TrimSpace(message.Title)
	body = strings.TrimSpace(body)
	if title != "" && body != "" {
		return title + "\n" + body
	}
	if title != "" {
		return title
	}
	return body
}

var (
	cmccCodeBlockRe  = regexp.MustCompile("(?s)```.*?```")
	cmccImageRe      = regexp.MustCompile(`!\[([^\]]*)\]\([^)]+\)`)
	cmccLinkRe       = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)
	cmccBoldRe       = regexp.MustCompile(`\*\*([^*]+)\*\*`)
	cmccUnderBoldRe  = regexp.MustCompile(`__([^_]+)__`)
	cmccInlineCodeRe = regexp.MustCompile("`([^`]+)`")
	cmccStrikeRe     = regexp.MustCompile(`~~([^~]+)~~`)
	cmccItalicRe     = regexp.MustCompile(`\*([^*\n]+)\*`)
	cmccHeadingRe    = regexp.MustCompile(`(?m)^#{1,6}\s*`)
	cmccListItemRe   = regexp.MustCompile(`(?m)^[\t ]*[-*+]\s+`)
	cmccNumberListRe = regexp.MustCompile(`(?m)^[\t ]*\d+\.\s+`)
	cmccQuoteRe      = regexp.MustCompile(`(?m)^>\s*`)
	cmccHrRe         = regexp.MustCompile(`(?m)^(-{3,}|\*{3,}|_{3,})\s*$`)
	cmccNewlinesRe   = regexp.MustCompile(`\n{3,}`)
)

// cmccPlainText 把 Markdown 降级为适合短信阅读的纯文本。
func cmccPlainText(text string) string {
	if text == "" {
		return text
	}
	text = cmccCodeBlockRe.ReplaceAllStringFunc(text, func(block string) string {
		return strings.TrimSpace(strings.Trim(block, "`"))
	})
	text = cmccImageRe.ReplaceAllString(text, "$1")
	text = cmccLinkRe.ReplaceAllString(text, "$1")
	text = cmccInlineCodeRe.ReplaceAllString(text, "$1")
	text = cmccBoldRe.ReplaceAllString(text, "$1")
	text = cmccUnderBoldRe.ReplaceAllString(text, "$1")
	text = cmccStrikeRe.ReplaceAllString(text, "$1")
	text = cmccItalicRe.ReplaceAllString(text, "$1")
	text = cmccHeadingRe.ReplaceAllString(text, "")
	text = cmccListItemRe.ReplaceAllString(text, "• ")
	text = cmccNumberListRe.ReplaceAllString(text, "• ")
	text = cmccQuoteRe.ReplaceAllString(text, "")
	text = cmccHrRe.ReplaceAllString(text, "")
	text = cmccNewlinesRe.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}

var cmccMediaExtensions = map[string]string{
	".jpg": "IMAGE", ".jpeg": "IMAGE", ".png": "IMAGE", ".gif": "IMAGE",
	".webp": "IMAGE", ".bmp": "IMAGE", ".heic": "IMAGE",
	".mp4": "VIDEO", ".mov": "VIDEO", ".avi": "VIDEO", ".mkv": "VIDEO", ".webm": "VIDEO",
	".mp3": "AUDIO", ".wav": "AUDIO", ".m4a": "AUDIO", ".aac": "AUDIO",
	".ogg": "AUDIO", ".amr": "AUDIO", ".flac": "AUDIO",
	".txt": "TEXT", ".md": "TEXT", ".log": "TEXT", ".csv": "TEXT",
	".pdf": "FILE", ".doc": "FILE", ".docx": "FILE", ".xls": "FILE", ".xlsx": "FILE",
	".ppt": "FILE", ".pptx": "FILE", ".zip": "FILE", ".rar": "FILE", ".7z": "FILE",
}

func inferCmccMediaType(rawURL string) (string, bool) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return "", false
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}
	mediaType, ok := cmccMediaExtensions[strings.ToLower(path.Ext(parsed.Path))]
	return mediaType, ok
}
