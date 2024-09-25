package src

import (
	"context"
	"fmt"
	"os"

	"whatsgpt/src/model"
	"whatsgpt/src/modules"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	openai "github.com/sashabaranov/go-openai"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
)


func createEventHandler(clt *model.MyClient) func(interface{}) {

    //clt.Chats = append(clt.Chats, modules.NoahChat(clt))
    clt.Modules = append(clt.Modules, modules.MensaKoeln(clt))

    return func(evt interface{}) {
        switch v := evt.(type) {
        case *events.Message:
            // skip my messages
            // if v.Info.IsFromMe {
            //  return
            // }

            for _, chat := range clt.Chats {
                chat.HandleMessage(v)
            }

            for _, module := range clt.Modules {
                module.HandleMessage(v)
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
    openai_client := openai.NewClient(openai_key)

    // create Chats

    // create client wrapper
    myclient := model.MyClient{
        WA: client,
        Openai: openai_client,
        Ctx: ctx,
        EventHandler: make([]uint32, 0),
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
    myclient.EventHandler = append(myclient.EventHandler, msgHandler )

    return client
}
