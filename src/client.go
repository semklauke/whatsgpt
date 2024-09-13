package whatsgpt

import (
	"context"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"

	"go.mau.fi/whatsmeow"
	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

	"github.com/bep/debounce"
	openai "github.com/sashabaranov/go-openai"
)

type MyClient struct {
	wa *whatsmeow.Client
	openai *openai.Client
	ctx context.Context
	eventHandler []uint32
	debouncer func(func())
	gptinstructions string
}

func createEventHandler(clt *MyClient) func(interface{}) {
	return func(evt interface{}) {
		switch v := evt.(type) {
		case *events.Message:
			// normal chat message
			if v.Info.IsFromMe {
				return
			}
			if v.IsEphemeral || v.Message.ImageMessage != nil {
				return
			}
			var messageBody = v.Message.GetConversation()
			if v.Info.Sender.User == "491711999899" {
				// noah
				fmt.Printf("[MSG from Noah]: %s\n", messageBody)
				clt.debouncer(func() {
					resp, err := clt.openai.CreateChatCompletion(
						context.Background(),
						openai.ChatCompletionRequest{
							Model: openai.GPT3Dot5Turbo,
							Messages: []openai.ChatCompletionMessage{
								{
									Role:    openai.ChatMessageRoleSystem,
									Content: clt.gptinstructions,
								},
								{
									Role:    openai.ChatMessageRoleUser,
									Content: messageBody,
								},
							},
						},
					)

					if err != nil {
						fmt.Printf("Error: %v\n", err)
						return
					}

					fmt.Printf("Ans: %s\n", resp.Choices[0].Message.Content)
					clt.wa.SendMessage(clt.ctx, v.Info.Chat, &waE2E.Message{
						Conversation: proto.String(resp.Choices[0].Message.Content),
					})
				})
			}
		}
	}
}

func CreateClient(ctx context.Context) *whatsmeow.Client {
	// logging for database and whatsapp client
	dbLog := waLog.Stdout("Database", "INFO", true)
	clientLog := waLog.Stdout("Client", "INFO", true)

	// create database
	container, err := sqlstore.New("sqlite3", "file:wastore.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}

	// create device store
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}

	// create new wa client
	client := whatsmeow.NewClient(deviceStore, clientLog)

	// create openai client
	openai_key := os.Getenv("OPENAI_KEY")
	if openai_key == "" {
		panic("'OPENAI_KEY' environment variable is not set or empty")
	}
	openai_client := openai.NewClient("")

	// load file with gtp instructions
	gptfile, err := os.ReadFile("src/chatgpt_instructions.txt")
    if err != nil {
        fmt.Print(err)
        panic(err)
    }
	// create client wrapper
	myclient := MyClient{
		wa: client,
		openai: openai_client,
		ctx: ctx,
		eventHandler: make([]uint32, 0),
		debouncer: debounce.New(70 * time.Second),
		gptinstructions: string(gptfile),
	}

	// is this a new login ?
	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(ctx)
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				// Render the QR code here
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		// Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	// add event listeners
	msgHandler := client.AddEventHandler(createEventHandler(&myclient))
	myclient.eventHandler = append(myclient.eventHandler, msgHandler )

	return client
}
