package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"

	"golang.design/x/clipboard"
)

type Config struct {
	VnextAddr                 string
	VnextPort                 int
	VnextUserID               string
	VnextUserAlterId          int
	StreamSettingNetwork      string
	StreamSettingSecurity     string
	StreamSettingTLSInsecure  bool
	StreamSettingWSPath       string
	StreamSettingWSHeaderHost string
}

var (
	vmessPrefix             = "vmess://"
	ErrorInvalidFormatVMess = errors.New("invalid format vmess")
)

func ParseConfigFastssh(input string) (*Config, error) {
	if !strings.Contains(input, vmessPrefix) {
		return nil, ErrorInvalidFormatVMess
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
	config := new(Config)

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

var rInput = flag.String("i", "", "Vmess string input")
var rOuput = flag.String("f", "./v2ray-config.json", "V2ray Config Json Path")

func main() {
	flag.Parse()

	err := clipboard.Init()
	if err != nil {
		panic(err)
	}
	clbContent := clipboard.Read(clipboard.FmtText)
	strContent := *rInput
	if len(strContent) <= 0 {
		strContent = string(clbContent)
	}
	cfg, err := ParseConfigFastssh(strContent)
	if err != nil {
		panic(err)
	}
	tmpl, err := template.New("vmessConfig").Parse(tmplString)
	if err != nil {
		panic(err)
	}
	f, err := os.OpenFile(*rOuput, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err := tmpl.Execute(f, cfg); err != nil {
		log.Println(err)
	}
}

var (
	tmplString = `{
  "inbounds": [{
    "port": 10808,
    "listen": "0.0.0.0",
    "protocol": "socks",
    "settings": {
      "udp": true
    }
  }],
  "outbounds": [{
    "protocol": "vmess",
    "settings": {
      "vnext": [{
        "address": "{{- .VnextAddr }}",
        "port": {{ .VnextPort }},
        "users": [{ "id": "{{ .VnextUserID }}", "alterId": {{ .VnextUserAlterId }} }]
      }]
    },
    "streamSettings": {
        "network": "{{ .StreamSettingNetwork }}",
        "security": "{{ .StreamSettingSecurity }}",
        "tlsSettings": {
            "allowInsecure": {{ .StreamSettingTLSInsecure }}
        },
        "wsSettings": {
        "path": "{{ .StreamSettingWSPath }}",
        "headers": {
            "Host": "{{ .StreamSettingWSHeaderHost }}"
        }
      }
    }
  },{
    "protocol": "freedom",
    "tag": "direct",
    "settings": {}
  }],
  "routing": {
    "domainStrategy": "IPOnDemand",
    "rules": [{
      "type": "field",
      "ip": ["geoip:private"],
      "outboundTag": "direct"
    }]
  }
}
`
)
