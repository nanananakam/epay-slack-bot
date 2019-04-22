package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/sclevine/agouti"
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
		SlackPost(cfg, epay())
	} else {
		SlackPost(cfg, "@"+postData.UserName+" お前は誰だ")
	}
}

func main() {
	log.Println("server startup......")
	http.HandleFunc("/v1/slack/inbound", BindSlackData)
	http.ListenAndServe(":80", nil)
}

func epay() string {
	var result string
	driver := agouti.ChromeDriver(agouti.ChromeOptions("args", []string{
		"--headless",
	}), agouti.Debug)

	err := driver.Start()
	if err != nil {
		driver.Stop()
		log.Println(err)
		return err.Error()
	}

	page, err := driver.NewPage(agouti.Browser("chrome"))
	if err != nil {
		driver.Stop()
		log.Println(err)
		return err.Error()
	}

	err = page.Navigate("https://prb01.payroll.co.jp/epayc/")
	if err != nil {
		driver.Stop()
		log.Println(err)
		return err.Error()
	}

	err = page.Find("input[name=\"copCd\"]").Fill(os.Getenv("ePayWorkCopCd"))
	if err != nil {
		driver.Stop()
		log.Println(err)
		return err.Error()
	}

	err = page.Find("input[name=\"empCd\"]").Fill(os.Getenv("ePayWorkEmpCd"))
	if err != nil {
		driver.Stop()
		log.Println(err)
		return err.Error()
	}

	err = page.Find("input[name=\"password\"]").Fill(os.Getenv("ePayWorkPassword"))
	if err != nil {
		driver.Stop()
		log.Println(err)
		return err.Error()
	}

	err = page.Find("button[type=\"submit\"]").Click()
	if err != nil {
		driver.Stop()
		log.Println(err)
		return err.Error()
	}

	err = page.Navigate("https://prb01.payroll.co.jp/epayc/mainPersonal.do?op=doSso&fwdSyscd=work&concd=calendar")
	if err != nil {
		driver.Stop()
		log.Println(err)
		return err.Error()
	}

	now := time.Now()

	fromValue, err := page.All(".work-control.work-calendar").At(now.Day() - 1).All("input[type=\"text\"]").At(0).Attribute("value")
	if err != nil {
		driver.Stop()
		log.Println(err)
		return err.Error()
	}

	mmss := fmt.Sprintf("%02d%02d", now.Hour(), now.Minute())
	if fromValue == "" {
		err = page.All(".work-control.work-calendar").At(now.Day() - 1).All("input[type=\"text\"]").At(0).Fill(mmss)
		if err != nil {
			driver.Stop()
			log.Println(err)
			return err.Error()
		}
		result = "出勤打刻 " + mmss
	} else {
		err = page.All(".work-control.work-calendar").At(now.Day() - 1).All("input[type=\"text\"]").At(1).Fill(mmss)
		if err != nil {
			driver.Stop()
			log.Println(err)
			return err.Error()
		}
		result = "退勤打刻 " + mmss
	}

	err = page.Find("#contentsRight > div.wrapperCenter.mt10 > a.buttonLBright.lastchild").Click()
	if err != nil {
		driver.Stop()
		log.Println(err)
		return err.Error()
	}

	err = page.Find("#navigation > ul > li.lastchild").Click()

	if err != nil {
		driver.Stop()
		log.Println(err)
		return err.Error()
	}

	err = driver.Stop()
	if err != nil {
		log.Println(err)
		return err.Error()
	}
	return result
}
