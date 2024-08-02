package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/handlerfunc"
	"github.com/kento13410/zoom_line_bot/zoom"
	"github.com/line/line-bot-sdk-go/linebot"
)

func main() {
	adapter := handlerfunc.New(callbackHandler)
	lambda.Start(adapter.ProxyWithContext)
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
