package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"whatsgpt/src"

	"go.mau.fi/whatsmeow"
)

func main() {

	// Check each hour if the client is still logged in
	keepAliveTimer := time.NewTicker(1 * time.Hour)
    defer keepAliveTimer.Stop()

    // Set up a channel to listen for system signals
    quitChannel := make(chan os.Signal, 1)
    signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)

    // create whatsapp client
    client := whatsgpt.CreateClient(context.Background())

    // run forever
    for {
        select {
        case <-keepAliveTimer.C:
            keepAliveCheck(client)
        case <-quitChannel:
            fmt.Println("Shutting down daemon...")
            client.Disconnect()
            return
        }
    }

}

func keepAliveCheck(client *whatsmeow.Client) {
	if !client.IsLoggedIn() {
		fmt.Println("WA client got disconnected. Someone should repair that (TODO)")
		// TODO: restore connection. Hint: look at https://pkg.go.dev/go.mau.fi/whatsmeow#Client.PairPhone
	}
}