package channel

import (
	"message-pusher/model"
	"testing"
)

func TestParseMultiSubChannels(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		wantLen int
		wantErr bool
	}{
		{
			name:    "corp + cmcc",
			raw:     `[{"type":"corp","url":"https://example.com/hook"},{"type":"cmcc","secret":"ak_1","account_id":"R1"}]`,
			wantLen: 2,
		},
		{
			name:    "single sub channel",
			raw:     `[{"type":"bark","secret":"key"}]`,
			wantLen: 1,
		},
		{name: "empty string", raw: "", wantErr: true},
		{name: "invalid json", raw: `not json`, wantErr: true},
		{name: "not an array", raw: `{"type":"corp"}`, wantErr: true},
		{name: "empty array", raw: `[]`, wantErr: true},
		{name: "missing type", raw: `[{"url":"https://example.com"}]`, wantErr: true},
		{name: "nested multi", raw: `[{"type":"multi"}]`, wantErr: true},
		{name: "nested group", raw: `[{"type":"group"}]`, wantErr: true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := parseMultiSubChannels(c.raw)
			if c.wantErr {
				if err == nil {
					t.Fatal("期望返回错误，实际成功")
				}
				return
			}
			if err != nil {
				t.Fatalf("不期望错误：%v", err)
			}
			if len(got) != c.wantLen {
				t.Fatalf("子渠道数量 = %d，期望 %d", len(got), c.wantLen)
			}
		})
	}

	subs, err := parseMultiSubChannels(`[{"type":"  corp  ","url":"https://example.com/hook","secret":"s"},{"type":"cmcc","secret":"ak_2","account_id":"R2"}]`)
	if err != nil {
		t.Fatal(err)
	}
	if subs[0].Type != "corp" {
		t.Errorf("type 未去空格：%q", subs[0].Type)
	}
	if subs[0].URL != "https://example.com/hook" || subs[0].Secret != "s" {
		t.Errorf("字段映射错误：%+v", subs[0])
	}
	if subs[1].AccountId != "R2" {
		t.Errorf("account_id 映射错误：%q", subs[1].AccountId)
	}
}

func TestMultiSubChannelStatusEnabled(t *testing.T) {
	subs, err := parseMultiSubChannels(`[{"type":"corp"}]`)
	if err != nil {
		t.Fatal(err)
	}
	if subs[0].Status != 1 {
		t.Errorf("子通道 status 应为启用(1)，实际 %d", subs[0].Status)
	}
	var _ *model.Channel = subs[0]
}
