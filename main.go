package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/kento13410/zoom_line_bot/zoom"
	"github.com/line/line-bot-sdk-go/linebot"
	"log"
	"net/http"
	"os"
	"strings"
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
	log.Printf("Starting server on port %s", port)
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
					token, err := zoom.GetAccessToken(os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"), os.Getenv("ACCOUNT_ID"))
					if err != nil {
						log.Fatal(err)
					}
					response, err := zoom.CreateZoomMeeting(token.AccessToken)
					if err != nil {
						bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(fmt.Sprintf("zoom meetingの作成に失敗しました: %s", err.Error()))).Do()
						return
					}

					bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(fmt.Sprintf("zoom meetingの作成に成功しました: %s", response.JoinURL))).Do()
				}
			}
		}
	}
	// 200を返す
	w.WriteHeader(200)
}
