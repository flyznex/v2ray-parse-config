package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	V2rayCfgTemplatePath string
	V2rayCfgOutputPath   string
	ClassCfgTemplatePath string
	ClassCfgOutputPath   string
	ModeApi              bool
	TeleBotAllowUsers    []int64
}

func InitConfig() *Config {
	cfg := new(Config)
	cfg.V2rayCfgTemplatePath = *flag.String("r", "./assets/template/v2ray-config.tml", "V2ray config template")
	cfg.V2rayCfgOutputPath = *flag.String("f", "./v2ray-config.json", "V2ray Config Json Path")
	cfg.ClassCfgTemplatePath = *flag.String("t", "./assets/template/config.tmpl", "template config")
	cfg.ClassCfgOutputPath = *flag.String("c", "./config.yaml", "Clash config output path")
	cfg.ModeApi = *flag.Bool("m", true, "enable mod api")
	lstAllowedUsers := os.Getenv("TELEGRAM_USER_ALLOWED")
	if lst := *flag.String("a", "", "List allow userID telegram"); len(lst) > 0 {
		lstAllowedUsers = lst
	}
	for _, u := range strings.Split(lstAllowedUsers, ",") {
		if uid, err := strconv.ParseInt(u, 64, 64); err != nil {
			cfg.TeleBotAllowUsers = append(cfg.TeleBotAllowUsers, uid)
		}
	}
	return cfg
}
