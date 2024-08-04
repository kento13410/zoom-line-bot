package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

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
					response, err := zoom.CreateZoomMeeting(token.AccessToken, time.Now())
					if err != nil {
						bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(fmt.Sprintf("zoom meetingの作成に失敗しました: %s", err.Error()))).Do()
						return
					}

					bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(fmt.Sprintf("zoom meetingの作成に成功しました: %s", response.JoinURL))).Do()
				} else if match, _ := regexp.MatchString(`次回会議 (\d{1,2}月\d{1,2}日 \d{1,2}:\d{2})`, message.Text); match {
					re := regexp.MustCompile(`次回会議 (\d{1,2}月\d{1,2}日 \d{1,2}:\d{2})`)
					match := re.FindStringSubmatch(message.Text)
					if len(match) > 1 {
						layout := "2006年1月2日 15:04"
						fullDate := fmt.Sprintf("%d年%s", time.Now().Year(), match[1])
						meetingTime, err := time.ParseInLocation(layout, fullDate, time.Local)
						if err != nil {
							bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(fmt.Sprintf("日付の解析に失敗しました: %s", err.Error()))).Do()
							return
						}
						token, err := zoom.GetAccessToken(os.Getenv("CLIENT_ID"), os.Getenv("CLIENT_SECRET"), os.Getenv("ACCOUNT_ID"))
						if err != nil {
							log.Fatal(err)
						}
						response, err := zoom.CreateZoomMeeting(token.AccessToken, meetingTime)
						if err != nil {
							bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(fmt.Sprintf("zoom meetingの作成に失敗しました: %s", err.Error()))).Do()
							return
						}

						bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(fmt.Sprintf("次回会議の日程は%sです。\nこちらのリンクから参加してください: %s\n本メッセージをアナウンスしてください。", meetingTime.Format("2006年1月2日 15:04"), response.JoinURL))).Do()
					}
				}
			}
		}
	}
	// 200を返す
	w.WriteHeader(200)
}
