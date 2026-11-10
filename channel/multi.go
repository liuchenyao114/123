package channel

import (
	"encoding/json"
	"errors"
	"fmt"
	"message-pusher/common"
	"message-pusher/model"
	"strings"
)

// multiSubChannel 是内联的子渠道配置，整体以 JSON 数组形式存放在 multi 通道的 other 字段中。
// 每个元素等价于一个精简的 model.Channel，type 决定其余字段的语义（与单通道一致）。
type multiSubChannel struct {
	Type      string `json:"type"`
	Secret    string `json:"secret,omitempty"`
	AppId     string `json:"app_id,omitempty"`
	AccountId string `json:"account_id,omitempty"`
	URL       string `json:"url,omitempty"`
	Other     string `json:"other,omitempty"`
}

func (s multiSubChannel) toChannel() *model.Channel {
	return &model.Channel{
		Type:      s.Type,
		Secret:    s.Secret,
		AppId:     s.AppId,
		AccountId: s.AccountId,
		URL:       s.URL,
		Other:     s.Other,
		Status:    common.ChannelStatusEnabled,
	}
}

func parseMultiSubChannels(raw string) ([]*model.Channel, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("渠道配置为空，请在通道配置中勾选并填写至少一个渠道")
	}
	var subs []multiSubChannel
	if err := json.Unmarshal([]byte(raw), &subs); err != nil {
		return nil, fmt.Errorf("渠道配置必须是合法的 JSON 数组：%w", err)
	}
	if len(subs) == 0 {
		return nil, errors.New("至少需要一个推送渠道")
	}
	channels := make([]*model.Channel, 0, len(subs))
	for i, s := range subs {
		t := strings.TrimSpace(s.Type)
		if t == "" {
			return nil, fmt.Errorf("第 %d 个子渠道缺少 type", i+1)
		}
		// 防止嵌套自身造成无限递归；group 同样会递归调用 SendMessage
		if t == model.TypeMulti || t == model.TypeGroup {
			return nil, fmt.Errorf("第 %d 个子渠道不能是 multi 或 group", i+1)
		}
		c := s.toChannel()
		c.Type = t
		channels = append(channels, c)
	}
	return channels, nil
}

// SendMultiMessage 把一条消息依次投递到本通道内联配置的所有子渠道。
// 单个子渠道失败不影响其余子渠道继续发送；全部发完后汇总错误返回。
// 通道自身仍只有一个 token，对外表现为「一通道一码」。
func SendMultiMessage(message *model.Message, user *model.User, channel_ *model.Channel) error {
	subs, err := parseMultiSubChannels(channel_.Other)
	if err != nil {
		return err
	}
	originalTo, originalChannel := message.To, message.Channel
	defer func() {
		message.To, message.Channel = originalTo, originalChannel
	}()

	var failures []string
	for i, sub := range subs {
		// group 等实现会改写 message.To / message.Channel，每个子渠道前恢复原值
		message.To, message.Channel = originalTo, originalChannel
		if err = SendMessage(message, user, sub); err != nil {
			failures = append(failures, fmt.Sprintf("第 %d 个子渠道(%s)：%s", i+1, sub.Type, err.Error()))
		}
	}
	if len(failures) > 0 {
		return errors.New(strings.Join(failures, "；"))
	}
	return nil
}
