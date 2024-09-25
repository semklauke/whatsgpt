package model

import (
	"time"

    "github.com/bep/debounce"
    "go.mau.fi/whatsmeow/types/events"
)

type ChatModule struct {
	Myclient *MyClient
    Debouncer func(func())
    Message_cache []*events.Message
    Handle_message func(msg *events.Message)
}


func NewChatModule(myclt *MyClient, debounce_time time.Duration) *ChatModule {
    var module ChatModule
    module.Myclient = myclt
    module.Message_cache = make([]*events.Message, 0)
    module.Debouncer = debounce.New(debounce_time)
    return &module
}

func (module *ChatModule) HandleMessage(msg *events.Message) {
    // skip old messages
    if time.Since(msg.Info.Timestamp) >= time.Minute {
        return
    }
    // store all messages in the chat with this users
    module.Message_cache = append(module.Message_cache, msg)
	module.Debouncer(func() { module.Handle_message(msg) })
}

func (module *ChatModule) ClearMessageCache() {
    module.Message_cache = make([]*events.Message, 0)
}