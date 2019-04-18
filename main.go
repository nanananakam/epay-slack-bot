package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
)

type SlackMsg struct {
	Channel  string `json:"channel"`
	Username string `json:"username,omitempty"`
	Text     string `json:"text"`
	Parse    string `json:"parse"`
}

type Config struct {
	WebhookUrl string `json:"webhook_url"`
	TargetUser string `json:"target_user"`
}

func ReadConfig() (*Config, error) {
	var cfg Config
	cfg.WebhookUrl = os.Getenv("webhook_url")
	cfg.TargetUser = os.Getenv("target_user")
	log.Println("config:", cfg)

	return &cfg, nil
}

func SlackPost(cfg *Config, text string) error {

	var data SlackMsg
	data.Parse = "full"
	data.Text = text
	jsonData, err := json.Marshal(data)
	WebhookUrl := cfg.WebhookUrl
	log.Println("webhookUrl", WebhookUrl)

	log.Println("jsonData:", string(jsonData), err)

	resp, err := http.PostForm(WebhookUrl, url.Values{"payload": {string(jsonData)}})
	if err != nil {
		log.Println("post Form Error:", resp, err)
	}
	log.Println(resp.Status, resp.Body, " : ", err)
	return err
}

type SlackHookMesage struct {
	Token       string `json:"token"`
	TeamId      string `json:"team_id"`
	ChannelId   string `json:"channel_id"`
	ChannelName string `json:"channel_name"`
	Timestamp   string `json:"timestamp"`
	UserId      string `json:"user_id"`
	UserName    string `json:"user_name"`
	Text        string `json:"text"`
	TriggerWord string `json:"trigger_word"`
}

func BindSlackData(_ http.ResponseWriter, r *http.Request) {
	var postData SlackHookMesage
	postData.Token = r.FormValue("token")
	postData.TeamId = r.FormValue("team_id")
	postData.ChannelId = r.FormValue("channel_id")
	postData.ChannelName = r.FormValue("channel_name")
	postData.Timestamp = r.FormValue("timestamp")
	postData.UserId = r.FormValue("user_id")
	postData.UserName = r.FormValue("user_name")
	postData.Text = r.FormValue("text")
	postData.TriggerWord = r.FormValue("trigger_word")

	log.Println(postData)

	log.Println("text: " + postData.Text)
	cfg, err := ReadConfig()
	if err != nil {
		log.Println("ReadConfig error! ", err)
		return
	}

	log.Println("config setting:", cfg.WebhookUrl)
	if postData.UserName == cfg.TargetUser {
		SlackPost(cfg, "はい")
	} else {
		SlackPost(cfg, "@"+postData.UserName+" お前は誰だ")
	}
}

func main() {
	log.Println("server startup......")
	http.HandleFunc("/v1/slack/inbound", BindSlackData)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
