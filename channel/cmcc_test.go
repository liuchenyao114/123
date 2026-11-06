package channel

import (
	"message-pusher/model"
	"strings"
	"testing"
)

func TestCmccPlainText(t *testing.T) {
	in := "# 标题\n\n**粗体** 与 [链接](https://example.com) 和 `code`\n\n- 项目一\n- 项目二\n\n> 引用"
	got := cmccPlainText(in)
	for _, want := range []string{"标题", "粗体", "链接", "code", "• 项目一", "引用"} {
		if !strings.Contains(got, want) {
			t.Errorf("期望包含 %q，实际结果：\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"**", "](https://", "# 标题", "- 项目一", "> 引用"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("不应包含 %q，实际结果：\n%s", unwanted, got)
		}
	}
	if cmccPlainText("") != "" {
		t.Error("空字符串应原样返回")
	}
	if cmccPlainText("普通文本") != "普通文本" {
		t.Error("无 Markdown 标记的文本不应被改写")
	}
}

func TestInferCmccMediaType(t *testing.T) {
	cases := []struct {
		url  string
		want string
		ok   bool
	}{
		{"https://example.com/a.png?x=1", "IMAGE", true},
		{"https://example.com/clip.MP4", "VIDEO", true},
		{"https://example.com/voice.mp3", "AUDIO", true},
		{"https://example.com/report.pdf", "FILE", true},
		{"https://example.com/notes.txt", "TEXT", true},
		{"https://example.com/message/uuid", "", false},
		{"", "", false},
	}
	for _, c := range cases {
		got, ok := inferCmccMediaType(c.url)
		if got != c.want || ok != c.ok {
			t.Errorf("inferCmccMediaType(%q) = (%q, %v)，期望 (%q, %v)", c.url, got, ok, c.want, c.ok)
		}
	}
}

func TestBuildCmccContent(t *testing.T) {
	cases := []struct {
		name string
		msg  *model.Message
		want string
	}{
		{"标题加正文并降级 Markdown", &model.Message{Title: "通知", Content: "**内容**"}, "通知\n内容"},
		{"正文为空时回退到描述", &model.Message{Description: "描述"}, "描述"},
		{"raw 模式保留原文", &model.Message{Title: "通知", Content: "**原样**", RenderMode: "raw"}, "通知\n**原样**"},
		{"只有标题", &model.Message{Title: "只有标题"}, "只有标题"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := buildCmccContent(c.msg); got != c.want {
				t.Errorf("buildCmccContent() = %q，期望 %q", got, c.want)
			}
		})
	}
}

func TestParseCmccConfig(t *testing.T) {
	ch := &model.Channel{
		Type:      "cmcc",
		Secret:    " ak_test ",
		AccountId: " RECIPIENT ",
		Other:     " https://example.com/hook ",
	}
	cfg := parseCmccConfig(ch)
	if cfg.apiKey != "ak_test" {
		t.Errorf("API Key 未去空格：%q", cfg.apiKey)
	}
	if cfg.serverURL != cmccDefaultServerURL {
		t.Errorf("空服务地址应使用默认值，实际：%q", cfg.serverURL)
	}
	if cfg.version != cmccDefaultVersion {
		t.Errorf("空版本应使用默认值，实际：%q", cfg.version)
	}
	if cfg.defaultRecipient != "RECIPIENT" {
		t.Errorf("默认接收人解析错误：%q", cfg.defaultRecipient)
	}
	if cfg.webhook != "https://example.com/hook" {
		t.Errorf("Webhook 解析错误：%q", cfg.webhook)
	}
	if cfg.key != "ak_test|"+cmccDefaultServerURL {
		t.Errorf("连接复用 key 错误：%q", cfg.key)
	}

	custom := &model.Channel{Secret: "k2", URL: " wss://custom.example/ws ", AppId: " 3.0 "}
	customCfg := parseCmccConfig(custom)
	if customCfg.serverURL != "wss://custom.example/ws" || customCfg.version != "3.0" {
		t.Errorf("自定义配置未生效：%+v", customCfg)
	}
}
