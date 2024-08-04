package zoom

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type CreateMeetingRequest struct {
	Topic     string   `json:"topic"`
	Type      int      `json:"type"`
	StartTime string   `json:"start_time"`
	Duration  int      `json:"duration"`
	Timezone  string   `json:"timezone"`
	Password  string   `json:"password"`
	Agenda    string   `json:"agenda"`
	Settings  Settings `json:"settings"`
}

type Settings struct {
	HostVideo        bool   `json:"host_video"`
	ParticipantVideo bool   `json:"participant_video"`
	JoinBeforeHost   bool   `json:"join_before_host"`
	MuteUponEntry    bool   `json:"mute_upon_entry"`
	Watermark        bool   `json:"watermark"`
	UsePmi           bool   `json:"use_pmi"`
	ApprovalType     int    `json:"approval_type"`
	RegistrationType int    `json:"registration_type"`
	Audio            string `json:"audio"`
	AutoRecording    string `json:"auto_recording"`
	EnforceLogin     bool   `json:"enforce_login"`
	AlternativeHosts string `json:"alternative_hosts"`
}

type CreateMeetingResponse struct {
	ID        int    `json:"id"`
	JoinURL   string `json:"join_url"`
	StartTime string `json:"start_time"`
	Topic     string `json:"topic"`
}

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

func GetAccessToken(clientID, clientSecret, accountID string) (*TokenResponse, error) {
	endpoint := "https://zoom.us/oauth/token"
	data := url.Values{}
	data.Set("grant_type", "account_credentials")
	data.Set("account_id", accountID)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	auth := base64.StdEncoding.EncodeToString([]byte(url.QueryEscape(clientID) + ":" + url.QueryEscape(clientSecret)))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var tokenResponse TokenResponse
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, err
	}

	return &tokenResponse, nil
}

func CreateZoomMeeting(token string, meetingTime time.Time) (*CreateMeetingResponse, error) {
	client := resty.New()

	meetingRequest := &CreateMeetingRequest{
		Topic:     "Business Contest Meeting",
		Type:      2, // Scheduled meeting
		StartTime: meetingTime.Format("2006-01-02T15:04:05"),
		Duration:  60, // 1 hour
		Timezone:  "JTC",
		Password:  "123456",
		Agenda:    "Discuss project status",
		Settings: Settings{
			HostVideo:        true,
			ParticipantVideo: true,
			JoinBeforeHost:   true,
			MuteUponEntry:    true,
			Watermark:        false,
			UsePmi:           false,
			ApprovalType:     0,
			RegistrationType: 1,
			Audio:            "both",
			AutoRecording:    "none",
			EnforceLogin:     false,
			AlternativeHosts: "",
		},
	}

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+token).
		SetHeader("Content-Type", "application/json").
		SetBody(meetingRequest).
		SetResult(&CreateMeetingResponse{}).
		Post("https://api.zoom.us/v2/users/me/meetings")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		return nil, fmt.Errorf("error: %s", resp.String())
	}

	return resp.Result().(*CreateMeetingResponse), nil
}
