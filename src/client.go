package whatsgpt

import (
    "context"
    "fmt"
    "os"

    _ "github.com/mattn/go-sqlite3"
    "github.com/mdp/qrterminal"
    "go.mau.fi/whatsmeow"
    "go.mau.fi/whatsmeow/store/sqlstore"
    "go.mau.fi/whatsmeow/types/events"
    waLog "go.mau.fi/whatsmeow/util/log"
    openai "github.com/sashabaranov/go-openai"
)

type MyClient struct {
    wa *whatsmeow.Client
    openai *openai.Client
    ctx context.Context
    eventHandler []uint32
    chats []*Chat
}

func createEventHandler(clt *MyClient) func(interface{}) {

    noah := NoahChat(clt)

    return func(evt interface{}) {
        switch v := evt.(type) {
        case *events.Message:
            // skip my messages
            // if v.Info.IsFromMe {
            //  return
            // }
            noah.HandleMessage(v)
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
    openai_client := openai.NewClient(openai_key)

    // create Chats

    // create client wrapper
    myclient := MyClient{
        wa: client,
        openai: openai_client,
        ctx: ctx,
        eventHandler: make([]uint32, 0),
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
