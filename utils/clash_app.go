package utils

import (
	"encoding/json"
	"html/template"
	"os"
	"time"
	"v2rayconfig/config"

	"github.com/rs/zerolog/log"
)

type ClassAppParser struct {
	ClassCfgTemplatePath string
	ClassCfgOutputPath   string
}

func NewClassAppParser(cfg *config.Config) Parser {
	return &ClassAppParser{
		ClassCfgTemplatePath: cfg.ClassCfgTemplatePath,
		ClassCfgOutputPath:   cfg.ClassCfgOutputPath,
	}
}

func parseConfigClash(input string) (map[string]interface{}, error) {
	out, err := parseVmessString(input)
	if err != nil {
		return nil, err
	}
	/* {
	  "type": "vmess",
	  "name": "sg-lb.vhax.net",
	  "ws-opts": {
	    "path": "/sshkit/03602ak019/6350f87e63f41/",
	    "headers": {
	      "host": "dl.kgvn.garenanow.com"
	    }
	  },
	  "server": "sg-lb.vhax.net",
	  "port": "443",
	  "uuid": "6fea1649-425b-4092-bf53-29792152c925",
	  "alterId": "0",
	  "cipher": "auto",
	  "network": "ws",
	  "tls": true
	}*/
	proxyCfg := map[string]interface{}{
		"type": "vmess",
		"name": "vmess",
		"ws-opts": map[string]interface{}{
			"path": out["path"],
			"headers": map[string]interface{}{
				"host": out["host"],
			},
		},
		"server":  out["add"],
		"port":    out["port"],
		"uuid":    out["id"],
		"alterId": out["aid"],
		"cipher":  "auto",
		"network": "ws",
	}

	if _, ok := out["tls"]; ok {
		proxyCfg["tls"] = true
		proxyCfg["skip-cert-verify"] = true
	}
	return proxyCfg, nil
}
func (p *ClassAppParser) GenConfig(input string) error {
	return genClashConfig(input, p.ClassCfgTemplatePath, p.ClassCfgOutputPath)
}
func (p *ClassAppParser) GetName() string {
	return "ClassAppParser"
}

func genClashConfig(input string, templatePath, storePath string) error {
	tmplC, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}
	cparse, err := parseConfigClash(input)
	if err != nil {
		return err
	}
	jsonConfig, err := json.Marshal(cparse)
	if err != nil {
		return err
	}
	fc, err := os.OpenFile(storePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	type cfg struct {
		Value     string
		UpdatedAt string
	}
	data := cfg{
		Value:     string(jsonConfig),
		UpdatedAt: time.Now().In(time.FixedZone("UTC+7", +7*60*60)).Format("2006-01-02T15:04:05"),
	}
	if err := tmplC.Execute(fc, data); err != nil {
		log.Error().Err(err).Msg("")
	}
	defer fc.Close()
	return nil
}
