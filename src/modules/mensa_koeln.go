package modules

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"whatsgpt/src/model"

	//waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	//"google.golang.org/protobuf/proto"

	"github.com/PuerkitoBio/goquery"
)

func MensaKoeln(clt *model.MyClient) *model.ChatModule {
    mensa := model.NewChatModule(clt, 5 * time.Second)

    // handle messages...
    mensa.Handle_message = func(msg *events.Message) {
        // skip messages we don't want to handle
        if msg.IsEphemeral || msg.Message.ImageMessage != nil || msg.IsEdit || msg.Info.Chat.IsBroadcastList() {
            return
        }

        // check if one of the messages asks for mensa
        isMensaRequest := false
        for _, m := range mensa.Message_cache {
            if m.Message.GetConversation() == "mensa pls" {
                isMensaRequest = true
                break
            }
        }
        defer mensa.ClearMessageCache()
        
        // only continue if this is a mensa request
        if !isMensaRequest {
            return
        }


        // Request the HTML page
        res, err := http.Get("https://koeln.my-mensa.de/essen.php?v=14393974&hyp=1&lang=de&mensa=um#um_tag_2024273")
        defer res.Body.Close()
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            return
        }


        if res.StatusCode != 200 {
            fmt.Printf("status code error: %d %s", res.StatusCode, res.Status)
            return
        }

        // Load the HTML document
        doc, err := goquery.NewDocumentFromReader(res.Body)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
        }

        ans := ""
        removeFoodWarnings := regexp.MustCompile(`\([\d,\D]*\)`)

        doc.Find(".checkgroupdividers.ui-listview.ui-group-theme-a li h3.ct.ui-li-heading.text2share").Each(func(i int, s *goquery.Selection) {
            if i == 0 {
                return
            }

            ans += strings.TrimSpace(removeFoodWarnings.ReplaceAllString(s.Text(), "")) + "\n"
        })

        // respond to message
        fmt.Printf("Ans: %s\n", ans)
        // clt.WA.SendMessage(clt.Ctx, msg.Info.Chat, &waE2E.Message{
        //     Conversation: proto.String(ans),
        // })
    }

    return mensa
}