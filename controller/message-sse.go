package controller

import (
	"github.com/gin-gonic/gin"
	"io"
	"message-pusher/model"
	"sync"
)

var messageChanBufferSize = 10

var messageChanStore struct {
	Map   map[int]*chan *model.Message
	Mutex sync.RWMutex
}

func messageChanStoreAdd(messageChan *chan *model.Message, userId int) {
	messageChanStore.Mutex.Lock()
	defer messageChanStore.Mutex.Unlock()
	messageChanStore.Map[userId] = messageChan
}

func messageChanStoreRemove(userId int) {
	messageChanStore.Mutex.Lock()
	defer messageChanStore.Mutex.Unlock()
	delete(messageChanStore.Map, userId)
}

func init() {
	messageChanStore.Map = make(map[int]*chan *model.Message)
}

func syncMessageToUser(message *model.Message, userId int) {
	messageChanStore.Mutex.RLock()
	defer messageChanStore.Mutex.RUnlock()
	messageChan, ok := messageChanStore.Map[userId]
	if !ok {
		return
	}
	// 非阻塞发送：客户端消费不过来时丢弃旧消息，既不会阻塞推送主流程，
	// 也不会在持锁状态下等待，避免与 Remove 抢写锁造成死锁。
	select {
	case *messageChan <- message:
	default:
	}
}

func GetNewMessages(c *gin.Context) {
	userId := c.GetInt("id")
	messageChan := make(chan *model.Message, messageChanBufferSize)
	messageChanStoreAdd(&messageChan, userId)
	c.Stream(func(w io.Writer) bool {
		if msg, ok := <-messageChan; ok {
			c.SSEvent("message", *msg)
			return true
		}
		return false
	})
	messageChanStoreRemove(userId)
	close(messageChan)
}
