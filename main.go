package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/zoom-lib-golang/zoom-lib-golang"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	bot, err := linebot.New(
		os.Getenv("LINE_BOT_CHANNEL_SECRET"),
		os.Getenv("LINE_BOT_CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}
	events, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if strings.ToLower(message.Text) == "zoom" {
					zoomLink, err := createZoomMeeting()
					if err != nil {
						bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("Zoomの作成に失敗しました")).Do()
					} else {
						bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(fmt.Sprintf("ホスト用URL: %s\\n参加者用URL: %s", zoomLink.StartURL, zoomLink.JoinURL))).Do()
					}
				}
			}
		}
	}
}

func createZoomMeeting() (zoom.Meeting, error) {
	apiKey := os.Getenv("ZOOM_API_KEY")
	apiSecret := os.Getenv("ZOOM_API_SECRET")
	userId := os.Getenv("USER_ID")

	client := zoom.NewClient(apiKey, apiSecret)
	return client.CreateMeeting(zoom.CreateMeetingOptions{
		HostID: userId,
	})
}
