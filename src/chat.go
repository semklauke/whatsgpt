package whatsgpt

import (
	"time"

	"github.com/bep/debounce"
	"go.mau.fi/whatsmeow/types/events"
)

type Chat struct {
	myclient *MyClient
	userid string
	debouncer func(func())
	gptinstructions string
	message_cache []*events.Message
	handle_message func(msg *events.Message)
}

func NewChat(myclt *MyClient, userid string, debounce_time time.Duration) *Chat {
	var chat Chat
	chat.myclient = myclt
	chat.userid = userid
	chat.message_cache = make([]*events.Message, 0)
	chat.debouncer = debounce.New(debounce_time)
	return &chat
}

func (chat *Chat) HandleMessage(msg *events.Message) {
	chat.message_cache = append(chat.message_cache, msg)
	if msg.Info.Chat.User == chat.userid {
		chat.debouncer(func() { chat.handle_message(msg) })
	}
}