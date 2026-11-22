package channel

import (
	"errors"
	"fmt"
	"message-pusher/common"
	"message-pusher/model"
	"strings"
)

func SendGroupMessage(message *model.Message, user *model.User, channel_ *model.Channel) error {
	subChannels := strings.Split(channel_.AppId, "|")
	var subTargets []string
	if message.To != "" {
		subTargets = strings.Split(message.To, "|")
	} else {
		subTargets = strings.Split(channel_.AccountId, "|")
	}
	if len(subChannels) != len(subTargets) {
		return errors.New("无效的群组消息配置，子通道数量与子目标数量不一致")
	}
	errMessage := ""
	for i := 0; i < len(subChannels); i++ {
		// 每个子通道使用独立的消息副本，避免原地修改共享的 message 导致串扰
		subMessage := *message
		subMessage.To = subTargets[i]
		subMessage.Channel = subChannels[i]
		subChannel, err := model.GetChannelByName(subChannels[i], user.Id)
		if err != nil {
			return errors.New("获取群组消息子通道失败：" + err.Error())
		}
		if subChannel.Type == model.TypeGroup {
			return errors.New("群组消息子通道不能是群组消息")
		}
		if subChannel.Status != common.ChannelStatusEnabled {
			errMessage += fmt.Sprintf("发送群组消息子通道 %s 失败：该通道已被禁用\n", subChannels[i])
			continue
		}
		err = SendMessage(&subMessage, user, subChannel)
		if err != nil {
			errMessage += fmt.Sprintf("发送群组消息子通道 %s 失败：%s\n", subChannels[i], err.Error())
		}
	}
	if errMessage != "" {
		return errors.New(errMessage)
	}
	return nil
}
