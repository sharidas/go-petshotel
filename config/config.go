package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Configs struct {
	Configs []Configuration `json: "Configurations"`
}

type Configuration struct {
	MongoURL   string `json: "MongoURI"`
	MailSender string `json: "MailSender"`
	MailTo     string `json: "MailTo"`
	MailPort   int    `json: "MailPort"`
	MailServer string `json: "MailServer"`
	MailPasswd string `json: "MailPasswd"`
	AppPort    int    `json: "AppPort"`
}

func (c *Configuration) ConfigParser() error {
	jsonByte, err := ioutil.ReadFile("conf.json")

	if err != nil {
		fmt.Println("Error: ", err)
	}

	err1 := json.Unmarshal(jsonByte, c)

	if err1 != nil {
		fmt.Println("Error unmarshal data: ", err1)
	}

	return err
}
