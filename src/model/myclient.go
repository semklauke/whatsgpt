package model

import (
    "context"
    
    "go.mau.fi/whatsmeow"
    openai "github.com/sashabaranov/go-openai"
)

type MyClient struct {
    WA *whatsmeow.Client
    Openai *openai.Client
    Ctx context.Context
    EventHandler []uint32
    Modules []*ChatModule
    Chats []*UserChat
}