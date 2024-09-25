package model

import (
	"time"

	"github.com/bep/debounce"
	"go.mau.fi/whatsmeow/types/events"
)

type UserChat struct {
    ChatModule
    Userid string
    Gptinstructions string
}

func NewUserChat(myclt *MyClient, userid string, debounce_time time.Duration) *UserChat {
    var chat UserChat
    chat.Myclient = myclt
    chat.Message_cache = make([]*events.Message, 0)
    chat.Debouncer = debounce.New(debounce_time)
    chat.Userid = userid
    return &chat
}

func (chat *UserChat) HandleMessage(msg* events.Message) {
    // skip old messages
    if time.Since(msg.Info.Timestamp) >= time.Minute {
        return
    }
    if msg.Info.Chat.User == chat.Userid {
        // only store message for this chat
        chat.Message_cache = append(chat.Message_cache, msg)
    }
    
    if msg.Info.Sender.User == chat.Userid {
        // answer when the other users writes
        chat.Debouncer(func() { chat.Handle_message(msg) })
    }
}