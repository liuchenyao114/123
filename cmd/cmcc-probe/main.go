// Command cmcc-probe 独立验证中国移动「新消息 ClawBot」网关的连通性，
// 不需要启动 message-pusher，也不需要数据库或前端构建产物。
//
// 交叉编译为 Linux amd64：
//
//	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o cmcc-probe-linux-amd64 ./cmd/cmcc-probe
//
// 常用用法：
//
//	CMCC_API_KEY=ak_xxx ./cmcc-probe-linux-amd64 -to RECIPIENT_ID -msg "测试消息"
//	./cmcc-probe-linux-amd64 -key ak_xxx -listen          # 只监听入站消息，用来获取接收人 ID
//	./cmcc-probe-linux-amd64 -key ak_xxx -to ID -title 标题 -msg 正文
//
// 退出码：0 = 链路可达且消息已提交；1 = 连接/认证/发送失败；2 = 参数错误。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const defaultServerURL = "wss://5gvas01.cmicmaap.com/gtw-ai/openclaw/ws/msg"

type frame map[string]interface{}

type outgoing struct {
	Type      string `json:"type"`
	APIKey    string `json:"apiKey,omitempty"`
	Version   string `json:"version,omitempty"`
	To        string `json:"to,omitempty"`
	Content   string `json:"content,omitempty"`
	MessageID string `json:"messageId,omitempty"`
}

func main() {
	apiKey := flag.String("key", os.Getenv("CMCC_API_KEY"), "CMCC API Key（也可用环境变量 CMCC_API_KEY）")
	serverURL := flag.String("url", defaultServerURL, "CMCC WebSocket 服务地址")
	version := flag.String("version", "2.0", "协议版本")
	to := flag.String("to", "", "接收人 ID；留空时先等待一条入站消息自动获取")
	title := flag.String("title", "", "消息标题")
	message := flag.String("msg", "message-pusher CMCC 通道测试消息", "消息正文")
	listen := flag.Bool("listen", false, "仅监听入站消息、不发送，用于获取接收人 ID")
	timeout := flag.Duration("timeout", 15*time.Second, "握手与认证超时")
	wait := flag.Duration("wait", 60*time.Second, "等待入站消息的时长")
	flag.Parse()

	if strings.TrimSpace(*apiKey) == "" {
		fmt.Fprintln(os.Stderr, "参数错误：缺少 API Key，请用 -key 或环境变量 CMCC_API_KEY 提供")
		os.Exit(2)
	}
	*apiKey = strings.TrimSpace(*apiKey)

	fmt.Println("=== CMCC 5G 消息探针（linux/amd64）===")
	fmt.Println("服务地址 :", *serverURL)
	fmt.Println("协议版本 :", *version)
	fmt.Println("API Key  :", mask(*apiKey))
	fmt.Println()

	header := http.Header{}
	header.Set("X-API-Key", *apiKey)
	dialer := websocket.Dialer{HandshakeTimeout: *timeout}
	conn, resp, err := dialer.Dial(*serverURL, header)
	if err != nil {
		if resp != nil {
			fail("WebSocket 握手失败（HTTP %d）：%v", resp.StatusCode, err)
		}
		fail("WebSocket 握手失败：%v", err)
	}
	defer func() { _ = conn.Close() }()
	fmt.Println("[1/4] WebSocket 握手成功")

	var writeMutex sync.Mutex
	write := func(v interface{}) error {
		writeMutex.Lock()
		defer writeMutex.Unlock()
		_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
		return conn.WriteJSON(v)
	}

	frames := make(chan frame, 64)
	go readFrames(conn, frames)

	done := make(chan struct{})
	defer close(done)
	go heartbeat(write, done)

	auth := outgoing{Type: "auth", APIKey: *apiKey, Version: *version}
	printOutgoing(auth)
	if err = write(auth); err != nil {
		fail("auth 帧发送失败：%v", err)
	}
	if err = waitForAuth(frames, *timeout); err != nil {
		fail("%v", err)
	}
	fmt.Println("[2/4] 认证成功（auth_ok）")

	if *listen {
		fmt.Println("监听入站消息中，请在手机「新消息 ClawBot」中向该应用号发送一条消息（按 Ctrl+C 可退出）…")
		recipient, err := waitForRecipient(frames, *wait)
		if err != nil {
			fail("%v", err)
		}
		fmt.Println("[✓] 检测到接收人 ID：", recipient)
		fmt.Println("    可用 ./cmcc-probe-linux-amd64 -key <API_KEY> -to", recipient, "-msg 测试 进行推送")
		return
	}

	recipient := strings.TrimSpace(*to)
	if recipient == "" {
		fmt.Println("未指定 -to，等待入站消息自动获取接收人，请在手机「新消息 ClawBot」中发送一条消息…")
		recipient, err = waitForRecipient(frames, *wait)
		if err != nil {
			fail("%v", err)
		}
		fmt.Println("[✓] 自动获取接收人：", recipient)
	}
	fmt.Println("[3/4] 接收人：", recipient)

	content := strings.TrimSpace(*message)
	if strings.TrimSpace(*title) != "" {
		content = strings.TrimSpace(*title) + "\n" + content
	}
	if content == "" {
		fail("消息内容不能为空")
	}
	send := outgoing{
		Type:      "send",
		APIKey:    *apiKey,
		To:        recipient,
		Content:   content,
		MessageID: fmt.Sprintf("probe_%d", time.Now().UnixMilli()),
	}
	printOutgoing(send)
	if err = write(send); err != nil {
		fail("send 帧发送失败：%v", err)
	}
	fmt.Println("[4/4] send 帧已写入 WebSocket，messageId =", send.MessageID)

	collectResponses(frames, 5*time.Second)
	fmt.Println()
	fmt.Println("=== 结论：链路可达，消息已提交给 CMCC 网关 ===")
	fmt.Println("注意：写入成功只代表服务端已接收，不代表手机端已送达或已读，请检查手机「新消息」应用号。")
}

func readFrames(conn *websocket.Conn, frames chan<- frame) {
	defer close(frames)
	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			select {
			case frames <- frame{"type": "__error__", "message": err.Error()}:
			default:
			}
			return
		}
		var f frame
		if err = json.Unmarshal(data, &f); err != nil {
			f = frame{"type": "__raw__", "message": string(data)}
		}
		select {
		case frames <- f:
		default:
		}
	}
}

func heartbeat(write func(interface{}) error, done <-chan struct{}) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			_ = write(outgoing{Type: "ping"})
		}
	}
}

func waitForAuth(frames <-chan frame, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case f, ok := <-frames:
			if !ok {
				return fmt.Errorf("连接在认证完成前关闭")
			}
			switch frameType(f) {
			case "auth_ok":
				return nil
			case "auth_failed":
				return fmt.Errorf("认证失败：%s（请检查 API Key）", firstString(f, "message"))
			case "__error__":
				return fmt.Errorf("WebSocket 读取失败：%s", firstString(f, "message"))
			default:
				printFrame(f)
			}
		case <-timer.C:
			return fmt.Errorf("等待 auth_ok 超时（%s）", timeout)
		}
	}
}

func waitForRecipient(frames <-chan frame, timeout time.Duration) (string, error) {
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	for {
		select {
		case f, ok := <-frames:
			if !ok {
				return "", fmt.Errorf("连接已关闭")
			}
			t := frameType(f)
			if t == "__error__" {
				return "", fmt.Errorf("WebSocket 读取失败：%s", firstString(f, "message"))
			}
			printFrame(f)
			sender := firstString(f, "from", "phone")
			if sender != "" && (t == "text_message" || t == "media_message" || t == "message") {
				return sender, nil
			}
		case <-timer.C:
			return "", fmt.Errorf("等待入站消息超时（%s）", timeout)
		}
	}
}

func collectResponses(frames <-chan frame, d time.Duration) {
	fmt.Printf("等待服务端回包（%s）…\n", d)
	timer := time.NewTimer(d)
	defer timer.Stop()
	for {
		select {
		case f, ok := <-frames:
			if !ok {
				return
			}
			printFrame(f)
		case <-timer.C:
			return
		}
	}
}

func frameType(f frame) string {
	s, _ := f["type"].(string)
	return s
}

func firstString(f frame, keys ...string) string {
	for _, key := range keys {
		if s, ok := f[key].(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}

func mask(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + "***"
}

func printOutgoing(v interface{}) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	var f frame
	if err = json.Unmarshal(b, &f); err != nil {
		return
	}
	if key, ok := f["apiKey"].(string); ok {
		f["apiKey"] = mask(key)
	}
	b, _ = json.Marshal(f)
	fmt.Println("→", string(b))
}

func printFrame(f frame) {
	if key, ok := f["apiKey"].(string); ok {
		f["apiKey"] = mask(key)
	}
	b, _ := json.Marshal(f)
	fmt.Println("←", string(b))
}

func fail(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "\n失败："+format+"\n", args...)
	os.Exit(1)
}
