package modules

import (
    "fmt"
    "os"
    "strings"
    "time"
    
    "whatsgpt/src/model"

    openai "github.com/sashabaranov/go-openai"
    waE2E "go.mau.fi/whatsmeow/proto/waE2E"
    "go.mau.fi/whatsmeow/types/events"
    "google.golang.org/protobuf/proto"
)

func NoahChat(clt *model.MyClient) *model.UserChat {
    noah := model.NewUserChat(clt, "491711999899", 90 * time.Second)

    // load file with gpt instructions
    gptfile, err := os.ReadFile("gptinstructions/noah.txt")
    if err != nil {
        panic(err)
    }
    noah.Gptinstructions = string(gptfile)

    // handle messages...
    noah.Handle_message = func(msg *events.Message) {
        // skip messages we don't want to handle
        if msg.IsEphemeral || msg.Message.ImageMessage != nil || msg.IsEdit || msg.Info.IsGroup || msg.Info.Chat.IsBroadcastList() {
            return
        }

        messageBody := msg.Message.GetConversation()
        fmt.Printf("[MSG from Noah]: %s\n", messageBody)

        // --- start of the openai intergation ---
        // create messages we want to send to chatgpt
        var openai_messages = []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleSystem,
                Content: noah.Gptinstructions,
            },
        }
        for _, cacheMsg := range noah.Message_cache {
            if body := cacheMsg.Message.GetConversation(); body != "" {
                if cacheMsg.Info.IsFromMe {
                    body = "Sem: " + body
                } else {
                    body = "Noah: " + body
                }
                openai_messages = append(openai_messages, openai.ChatCompletionMessage{
                    Role: openai.ChatMessageRoleUser,
                    Content: body,
                })
            }
        }
        // send request to openai api
        resp, err := clt.Openai.CreateChatCompletion(
            clt.Ctx,
            openai.ChatCompletionRequest{
                Model: openai.GPT4,
                Messages: openai_messages,
            },
        )

        if err != nil {
            fmt.Printf("Error: %v\n", err)
            return
        }

        // clear message cache
        noah.Message_cache = make([]*events.Message, 0)

        ans := resp.Choices[0].Message.Content;

        ans = strings.Replace(ans, "Sem :", "", 1)
        ans = strings.Replace(ans, "Noah :", "", 1)

        // respond to message
        fmt.Printf("Ans: %s\n", ans)
        clt.WA.SendMessage(clt.Ctx, msg.Info.Chat, &waE2E.Message{
            Conversation: proto.String(ans),
        })
    }

    return noah
}