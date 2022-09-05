package conf

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var Conf Config

type Config struct {
	Filter    Filter     `yaml:"filter"`
	Dispatch  Dispatch   `yaml:"dispatch"`
	Receivers []Receiver `yaml:"receivers"`
	Email     Email      `yaml:"email"`
	Matrix    []Matrix   `yaml:"matrix"`
}

type Filter struct {
	Label Label `yaml:"label"`
}

type Dispatch struct {
	LabelExtractSender []map[string]string `yaml:"labelExtractSender"`
	LabelMatch         []LabelMatch        `yaml:"labelMatch"`
}

type LabelMatch struct {
	Receiver []string `yaml:"receiver"`
	Label    `yaml:",inline"`
}

type Label struct {
	Key         []string            `yaml:"key"`
	Value       []string            `yaml:"value"`
	Combination []map[string]string `yaml:"combination"`
}

type Receiver struct {
	Name    string   `yaml:"name"`
	Type    string   `yaml:"type"`
	Webhook string   `yaml:"webhook"`
	Token   string   `yaml:"token"`
	ChatID  string   `yaml:"chatID"`
	RoomID  string   `yaml:"roomID"`
	Sender  string   `yaml:"sender"`
	Email   []string `yaml:"email"`
}

type Email struct {
	Type   string `yaml:"type"`
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	User   string `yaml:"user"`
	Secret string `yaml:"secret"`
	Sender string `yaml:"sender"`
}

type Matrix struct {
	Name     string `yaml:"name"`
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

var path *string

func init() {
	path = flag.String("conf", "conf/config.yml", "configuration file path")
}

func InitConf(_ context.Context) {
	if !flag.Parsed() {
		flag.Parse()
	}
	data, err := os.ReadFile(*path)
	if err != nil {
		log.Fatalln("failed to open configuration file. ", err.Error())
	}
	if err = yaml.Unmarshal(data, &Conf); err != nil {
		log.Fatalf("failed to parse configuration file. err: %s\n", err.Error())
	}
	fmt.Printf("%+v\n", Conf)
	// check channel
	emailConf := false
	for _, c := range Conf.Receivers {
		switch c.Type {
		case "email":
			if len(c.Email) == 0 {
				log.Fatalln("email needs to configure receive email address")
			}
			emailConf = true
		case "slack":
			if len(c.Webhook) == 0 {
				log.Fatalln("slack needs to configure webhook")
			}
		case "telegram":
			if len(c.Webhook) == 0 && len(c.Token) == 0 {
				log.Fatalln("telegram needs to configure webhook or token")
			}
			if len(c.ChatID) == 0 {
				log.Fatalln("telegram needs to configure chatID")
			}
		case "element":
			if len(c.Sender) == 0 {
				log.Fatalln("element needs to configure sender")
			}
			if len(c.RoomID) == 0 {
				log.Fatalln("element needs to configure roomID")
			}
		}
	}
	if emailConf {
		if Conf.Email.Type != "sendgrid" && Conf.Email.Type != "smtp" {
			log.Fatalln("unsupported mail type")
		}
		if len(Conf.Email.Host) == 0 {
			log.Fatalln("email needs to configure host")
		}
		if Conf.Email.Port == 0 {
			log.Fatalln("email needs to configure port")
		}
	}
}
