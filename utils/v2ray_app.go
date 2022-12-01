package utils

import (
	"encoding/base64"
	"encoding/json"
	"html/template"
	"log"
	"os"
	"strconv"
	"strings"
	"v2rayconfig/config"
	"v2rayconfig/model"
)

type V2rayParser struct {
	V2rayCfgTemplatePath string
	V2rayCfgOutputPath   string
}

func NewV2rayParser(cfg *config.Config) Parser {
	return &V2rayParser{
		V2rayCfgTemplatePath: cfg.V2rayCfgTemplatePath,
		V2rayCfgOutputPath:   cfg.V2rayCfgOutputPath,
	}
}
func (p *V2rayParser) GetName() string {
	return "V2rayParser"
}

func (p *V2rayParser) GenConfig(input string) error {
	return genVmessConfig(input, p.V2rayCfgTemplatePath, p.V2rayCfgOutputPath)
}

func parseVmessString(input string) (map[string]interface{}, error) {
	if err := CheckVmessConfigValid(input); err != nil {
		return nil, err
	}
	// cInput := input[len(vmessPrefix):]
	decodeBytes, err := base64.StdEncoding.DecodeString(input[len(vmessPrefix):])

	if err != nil {
		return nil, err
	}
	out := map[string]interface{}{}
	if err := json.Unmarshal(decodeBytes, &out); err != nil {
		return nil, err
	}

	/*{
	  "add": "my.v2-ray.com",
	  "aid": "0",
	  "id": "afef92d5-b5b9-46aa-a982-3914fe62c1c5",
	  "host": "mobiedu.vn",
	  "net": "ws",
	  "path": "/fastssh/1208/62f5b158a6c55/",
	  "port": "443",
	  "ps": "my.v2-ray.com",
	  "tls": "tls",
	  "type": "none",
	  "v": "2"
		}*/
	return out, nil
}

func parseConfigV2ray(input string) (*model.V2rayConfig, error) {
	out, err := parseVmessString(input)
	if err != nil {
		return nil, err
	}
	config := new(model.V2rayConfig)

	if add, ok := out["add"]; ok {
		config.VnextAddr = add.(string)
	}
	if id, ok := out["id"]; ok {
		config.VnextUserID = id.(string)
	}
	if aid, ok := out["aid"]; ok {
		config.VnextUserAlterId, _ = strconv.Atoi(aid.(string))
	}
	if host, ok := out["host"]; ok {
		config.StreamSettingWSHeaderHost = host.(string)
	}
	if net, ok := out["net"]; ok {
		config.StreamSettingNetwork = net.(string)
	}
	if path, ok := out["path"]; ok {
		config.StreamSettingWSPath = path.(string)
	}
	if port, ok := out["port"]; ok {
		config.VnextPort, _ = strconv.Atoi(port.(string))
	}
	if tls, ok := out["tls"]; ok {
		config.StreamSettingSecurity = tls.(string)
		config.StreamSettingTLSInsecure = strings.Compare(tls.(string), "tls") == 0
	}
	return config, nil
}

func genVmessConfig(input string, templatePath, storePath string) error {
	cfg, err := parseConfigV2ray(input)
	if err != nil {
		return err
	}
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	fe, err := os.OpenFile(storePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(fe, cfg); err != nil {
		log.Println(err)
	}
	fe.Close()
	return nil
}
