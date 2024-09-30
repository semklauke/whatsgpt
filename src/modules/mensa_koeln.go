package modules

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"whatsgpt/src/model"

	waE2E "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"

	"github.com/PuerkitoBio/goquery"
)

func MensaKoeln(clt *model.MyClient) *model.ChatModule {
    mensa := model.NewChatModule(clt, 2 * time.Second)

    // handle messages...
    mensa.Handle_message = func(msg *events.Message) {
        // skip messages we don't want to handle
        if msg.IsEphemeral || msg.Message.ImageMessage != nil || msg.IsEdit {
            return
        }

        mensaRequest := regexp.MustCompile(`^(?i)mensa pl(s+)( \w+)?$`)
        var (
            mensa_id string
            mensa_date time.Time
        )

        // check if one of the messages asks for mensa
        for _, m := range mensa.Message_cache {
            matches := mensaRequest.FindAllStringSubmatch(m.Message.GetConversation(), 1)
            if len(matches) == 0 {
                continue
            }
            mensa_date = time.Now().AddDate(0,0, len(matches[0][1])-1)
            if matches[0][2] != "" {
                mensa_id = strings.TrimSpace(matches[0][2])
            } else {
                mensa_id = "um"
            }
            break
        }
        defer mensa.ClearMessageCache()
        
        // only continue if this is a mensa request
        if mensa_id == "" {
            return
        }


        mensa_tag := fmt.Sprintf("%s_tag_%d%d",  mensa_id, mensa_date.Year(), mensa_date.Local().YearDay()-1)
        mensa_url := fmt.Sprintf("https://koeln.my-mensa.de/essen.php?hyp=1&lang=de&mensa=%s#%s", mensa_id, mensa_tag)
        fmt.Println(mensa_url)

        // Request the HTML page
        res, err := http.Get(mensa_url)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            return
        }
        defer res.Body.Close()


        if res.StatusCode != 200 {
            fmt.Printf("status code error: %d %s", res.StatusCode, res.Status)
            return
        }

        // body, err := io.ReadAll(res.Body)
        // fmt.Println(string(body))

        // Load the HTML document
        doc, err := goquery.NewDocumentFromReader(res.Body)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
        }


        dishes := make([]string, 0, 8)

        removeFoodWarnings := regexp.MustCompile(`\([\d,\D]*\)`)
        removeFoodWarnings2 := regexp.MustCompile(`\s\(.*?\)`)

        doc.Find("h3.ct.ui-li-heading.text2share").Each(func(i int, s *goquery.Selection) {
            href, ok := s.Parent().Attr("href")
            if ok && strings.Contains(href, mensa_tag) {
                dish_name := strings.TrimSpace(removeFoodWarnings.ReplaceAllString(s.Text(), ""))
                dish_name = strings.ReplaceAll(dish_name, "\nauch als Beilage erhält­lich .", "")
                supp_raw := s.Parent().Find("p.ct.text2share:not(.next)").First().Text()
                supp := strings.TrimSpace(removeFoodWarnings2.ReplaceAllString(supp_raw, ""))
                if !strings.Contains(dish_name, "Selbst­bedienung") {
                    dishes = append(dishes, "- " + dish_name)
                    if supp != "" {
                        dishes = append(dishes, "> _" + supp + "_")
                    }
                }
            }
        })

        ans := fmt.Sprintf("*Mensa %s* (%02d.%02d):\n", mensaMap[mensa_id], mensa_date.Day(), mensa_date.Month())
        ans += strings.Join(dishes, "\n")

        // respond to message
        //fmt.Printf("Ans: %s\n", ans)
        clt.WA.SendMessage(clt.Ctx, msg.Info.Chat, &waE2E.Message{
            Conversation: proto.String(ans),
        })
    }

    return mensa
}


var mensaMap = map[string]string{
    "um": "Zülpicher",
    "rks": "Lindenthal",
    "sds": "Südstadt",
}