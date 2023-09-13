package conf

import (
	"context"
	"flag"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var Conf Config

type Config struct {
	Alert     Alert      `yaml:"alert"`
	Label     Label      `yaml:"label"`
	Dispatch  Dispatch   `yaml:"dispatch"`
	Receivers []Receiver `yaml:"receivers"`
	Email     Email      `yaml:"email"`
	Slack     []Slack    `yaml:"slack"`
	Discord   []Discord  `yaml:"discord"`
	Matrix    []Matrix   `yaml:"matrix"`
}

type Alert struct {
	Filter Filter `yaml:"filter"`
}

type ReplaceValue struct {
	Regex string `yaml:"regex"`
	Value string `yaml:"value"`
}

type Label struct {
	Exclude struct {
		Label LabelKV `yaml:"label"`
	} `yaml:"exclude"`
	Keep struct {
		Label LabelKV `yaml:"label"`
	} `yaml:"keep"`
	Replace struct {
		Label struct {
			Key   []ReplaceValue `yaml:"key"`
			Value []ReplaceValue `yaml:"value"`
		} `yaml:"label"`
	} `yaml:"replace"`
}

type Filter struct {
	Label LabelKV `yaml:"label"`
}

type Dispatch struct {
	LabelExtractSender []map[string]string `yaml:"labelExtractSender"`
	LabelMatch         []LabelMatch        `yaml:"labelMatch"`
}

type LabelMatch struct {
	Receiver []string `yaml:"receiver"`
	LabelKV  `yaml:",inline"`
}

type LabelKV struct {
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

type Discord struct {
	Name  string `yaml:"name"`
	Token string `yaml:"token"`
}

type Slack struct {
	Name  string `yaml:"name"`
	Token string `yaml:"token"`
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
	// check channel
	emailConf, matrixConf, slackConf := false, false, false
	for _, c := range Conf.Receivers {
		switch c.Type {
		case "email":
			if len(c.Email) == 0 {
				log.Fatalln("email needs to configure receive email address")
			}
			emailConf = true
		case "slack":
			if len(c.Webhook) == 0 && (len(c.ChatID) == 0 || len(c.Sender) == 0) {
				log.Fatalln("slack needs to configure webhook or chatID and Sender")
			}
			if len(c.ChatID) != 0 && len(c.Sender) != 0 {
				slackConf = true
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
			matrixConf = true
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
		if len(Conf.Email.Secret) == 0 {
			log.Fatalln("email needs to configure secret")
		}
	}
	if slackConf {
		if len(Conf.Slack) == 0 {
			log.Fatalln("needs to configure slack sender")
		}
	}
	if matrixConf {
		if len(Conf.Matrix) == 0 {
			log.Fatalln("needs to configure matrix")
		}
		for _, matrix := range Conf.Matrix {
			if len(matrix.Host) == 0 {
				log.Fatalln("matrix needs to configure host")
			}
			if len(matrix.User) == 0 {
				log.Fatalln("matrix needs to configure user")
			}
			if len(matrix.Password) == 0 {
				log.Fatalln("matrix needs to configure password")
			}
			if len(matrix.Name) == 0 {
				log.Fatalln("matrix needs to configure name")
			}
		}
	}
}
