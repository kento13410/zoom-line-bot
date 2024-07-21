package main

import (
	"encoding/json"
	"fmt"
	"github.com/donvito/zoom-go/zoomAPI/constants/meeting"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	zoom "github.com/donvito/zoom-go/zoomAPI"
	"github.com/line/line-bot-sdk-go/linebot"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
}

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
					token, err := getAccessToken(os.Getenv("ZOOM_ACCOUNT_ID"), os.Getenv("ZOOM_CLIENT_ID")+":"+os.Getenv("ZOOM_CLIENT_SECRET"))
					zoomResp, err := createZoomMeeting(token)
					if err != nil {
						bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("Zoomの作成に失敗しました")).Do()
					} else {
						bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(fmt.Sprintf("ホスト用URL: %s\\n参加者用URL: %s", zoomResp.StartUrl, zoomResp.JoinUrl))).Do()
					}
				}
			}
		}
	}
}

func getAccessToken(accountID, auth string) (string, error) {
	endpoint := "https://zoom.us/oauth/token"
	data := url.Values{}
	data.Set("grant_type", "account_credentials")
	data.Set("account_id", accountID)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var tokenResponse TokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", err
	}

	return tokenResponse.AccessToken, nil
}

func createZoomMeeting(token string) (zoom.CreateMeetingResponse, error) {
	userId := os.Getenv("USER_ID")

	client := zoom.NewClient(os.Getenv("ZOOM_API_URL"), token)
	return client.CreateMeeting(
		userId,
		"",
		meeting.MeetingTypeInstant,
		"",
		30,
		"",
		"Asia/Japan",
		"",
		"",
		nil,
		nil,
	)
}
