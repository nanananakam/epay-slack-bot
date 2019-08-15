package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
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

func BindSlackData(w http.ResponseWriter, r *http.Request) {
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
	if postData.Token != os.Getenv("token") {
		log.Println("Invalid token : " + postData.Token)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	log.Println("text: " + postData.Text)
	cfg, err := ReadConfig()
	if err != nil {
		log.Println("ReadConfig error! ", err)
		return
	}

	log.Println("config setting:", cfg.WebhookUrl)
	if postData.UserName == cfg.TargetUser {
		SlackPost(cfg, epay(postData.Text))
	} else {
		SlackPost(cfg, "@"+postData.UserName+" お前は誰だ")
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	log.Println("server startup......")
	http.HandleFunc("/v1/slack/inbound", BindSlackData)
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}

func epay(input string) string {
	var result string
	driver := agouti.ChromeDriver(agouti.ChromeOptions("args", []string{
		"--headless",
		"--no-sandbox",
		"--disable-gpu",
		"--window-size=1280,1024",
		"--disable-dev-shm-usage",
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

	time.Sleep(10 * time.Second)

	ePayWorkCopCd := os.Getenv("ePayWorkCopCd")
	ePayWorkEmpCd := os.Getenv("ePayWorkEmpCd")
	ePayWorkPassword := os.Getenv("ePayWorkPassword")

	log.Println(ePayWorkCopCd)
	log.Println(ePayWorkEmpCd)
	log.Println(ePayWorkPassword)

	err = page.Find("input[name=\"copCd\"]").Fill(ePayWorkCopCd)
	if err != nil {
		html, err2 := page.HTML()
		if err2 != nil {
			log.Println(err2)
		} else {
			log.Println(html)
		}
		driver.Stop()
		log.Println(err)
		return err.Error()
	}

	err = page.Find("input[name=\"empCd\"]").Fill(ePayWorkEmpCd)
	if err != nil {
		html, err2 := page.HTML()
		if err2 != nil {
			log.Println(err2)
		} else {
			log.Println(html)
		}
		driver.Stop()
		log.Println(err)
		return err.Error()
	}

	err = page.Find("input[name=\"password\"]").Fill(ePayWorkPassword)
	if err != nil {
		html, err2 := page.HTML()
		if err2 != nil {
			log.Println(err2)
		} else {
			log.Println(html)
		}
		driver.Stop()
		log.Println(err)
		return err.Error()
	}

	err = page.Find("button[type=\"submit\"]").Click()
	if err != nil {
		html, err2 := page.HTML()
		if err2 != nil {
			log.Println(err2)
		} else {
			log.Println(html)
		}
		driver.Stop()
		log.Println(err)
		return err.Error()
	}

	page.Session().SetImplicitWait(10000)

	err = page.Navigate("https://prb01.payroll.co.jp/epayc/mainPersonal.do?op=doSso&fwdSyscd=work&concd=calendar")
	if err != nil {
		html, err2 := page.HTML()
		if err2 != nil {
			log.Println(err2)
		} else {
			log.Println(html)
		}
		driver.Stop()
		log.Println(err)
		return err.Error()
	}

	page.Session().SetImplicitWait(10000)

	now := time.Now()

	if strings.Contains(input, "休日") {
		err = page.All(".work-control.work-calendar").At(now.Day() - 1).All("input[type=\"checkbox\"]").At(0).Check()
		if err != nil {
			html, err2 := page.HTML()
			if err2 != nil {
				log.Println(err2)
			} else {
				log.Println(html)
			}
			driver.Stop()
			log.Println(err)
			return err.Error()
		}
		err = page.All(".work-control.work-calendar").At(now.Day() - 1).All("input[type=\"text\"]").At(0).Fill("")
		if err != nil {
			html, err2 := page.HTML()
			if err2 != nil {
				log.Println(err2)
			} else {
				log.Println(html)
			}
			driver.Stop()
			log.Println(err)
			return err.Error()
		}
		err = page.All(".work-control.work-calendar").At(now.Day() - 1).All("input[type=\"text\"]").At(1).Fill("")
		if err != nil {
			html, err2 := page.HTML()
			if err2 != nil {
				log.Println(err2)
			} else {
				log.Println(html)
			}
			driver.Stop()
			log.Println(err)
			return err.Error()
		}
		result = "休日打刻"
	} else if strings.Contains(input, "有給") {
		err = page.All(".work-control.work-calendar").At(now.Day() - 1).All("input[type=\"checkbox\"]").At(0).Check()
		if err != nil {
			html, err2 := page.HTML()
			if err2 != nil {
				log.Println(err2)
			} else {
				log.Println(html)
			}
			driver.Stop()
			log.Println(err)
			return err.Error()
		}
		err = page.All(".work-control.work-calendar").At(now.Day() - 1).All("input[type=\"text\"]").At(0).Fill("")
		if err != nil {
			html, err2 := page.HTML()
			if err2 != nil {
				log.Println(err2)
			} else {
				log.Println(html)
			}
			driver.Stop()
			log.Println(err)
			return err.Error()
		}
		err = page.All(".work-control.work-calendar").At(now.Day() - 1).All("input[type=\"text\"]").At(1).Fill("")
		if err != nil {
			html, err2 := page.HTML()
			if err2 != nil {
				log.Println(err2)
			} else {
				log.Println(html)
			}
			driver.Stop()
			log.Println(err)
			return err.Error()
		}
		err = page.All(".work-control.work-calendar").At(now.Day() - 1).All("select").At(0).All("option[value=\"2000\"]").Click()
		if err != nil {
			html, err2 := page.HTML()
			if err2 != nil {
				log.Println(err2)
			} else {
				log.Println(html)
			}
			driver.Stop()
			log.Println(err)
			return err.Error()
		}
		result = "有給打刻"
	} else {
		fromValue, err := page.All(".work-control.work-calendar").At(now.Day() - 1).All("input[type=\"text\"]").At(0).Attribute("value")
		if err != nil {
			html, err2 := page.HTML()
			if err2 != nil {
				log.Println(err2)
			} else {
				log.Println(html)
			}
			driver.Stop()
			log.Println(err)
			return err.Error()
		}

		mmss := fmt.Sprintf("%02d%02d", now.Hour(), now.Minute())
		if err != nil {
			html, err2 := page.HTML()
			if err2 != nil {
				log.Println(err2)
			} else {
				log.Println(html)
			}
			driver.Stop()
			log.Println(err)
			return err.Error()
		}
		if fromValue == "" {
			err = page.All(".work-control.work-calendar").At(now.Day() - 1).All("input[type=\"text\"]").At(0).Fill(mmss)
			if err != nil {
				html, err2 := page.HTML()
				if err2 != nil {
					log.Println(err2)
				} else {
					log.Println(html)
				}
				driver.Stop()
				log.Println(err)
				return err.Error()
			}
			result = "出勤打刻 " + mmss
		} else {
			err = page.All(".work-control.work-calendar").At(now.Day() - 1).All("input[type=\"checkbox\"]").At(0).Check()
			if err != nil {
				html, err2 := page.HTML()
				if err2 != nil {
					log.Println(err2)
				} else {
					log.Println(html)
				}
				driver.Stop()
				log.Println(err)
				return err.Error()
			}
			err = page.All(".work-control.work-calendar").At(now.Day() - 1).All("input[type=\"text\"]").At(1).Fill(mmss)
			if err != nil {
				html, err2 := page.HTML()
				if err2 != nil {
					log.Println(err2)
				} else {
					log.Println(html)
				}
				driver.Stop()
				log.Println(err)
				return err.Error()
			}
			result = "退勤打刻 " + mmss
		}
	}

	err = page.Find("#contentsRight > div.wrapperCenter.mt10 > a.buttonLBright.lastchild").Click()
	if err != nil {
		html, err2 := page.HTML()
		if err2 != nil {
			log.Println(err2)
		} else {
			log.Println(html)
		}
		driver.Stop()
		log.Println(err)
		return err.Error()
	}

	err = page.Find("#navigation > ul > li.lastchild").Click()

	if err != nil {
		html, err2 := page.HTML()
		if err2 != nil {
			log.Println(err2)
		} else {
			log.Println(html)
		}
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
