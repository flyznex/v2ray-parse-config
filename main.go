package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/gin-gonic/gin"
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
	serverPort              = ":8080"
)

func parseVmess(input string) (map[string]interface{}, error) {
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
	return out, nil
}

func ParseConfigV2ray(input string) (*Config, error) {
	out, err := parseVmess(input)
	if err != nil {
		return nil, err
	}
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

func ParseConfigClash(input string) (map[string]interface{}, error) {
	out, err := parseVmess(input)
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

var rInput = flag.String("i", "", "Vmess string input")
var rOuput = flag.String("f", "./v2ray-config.json", "V2ray Config Json Path")
var clashCfgTemplatePath = flag.String("t", "./config.tmpl", "template config")
var cOutput = flag.String("c", "./config.yaml", "Clash config output path")
var modeApi = flag.Bool("m", true, "enable mod api")

func main() {
	flag.Parse()

	if *modeApi {
		if p := os.Getenv("PORT"); len(p) > 0 {
			serverPort = fmt.Sprintf(":%s", p)
		}
		r := gin.Default()
		r.GET("/clash-config", func(c *gin.Context) {
			c.FileAttachment(string(*cOutput), "config.yaml")
		})
		r.POST("/config", gin.BasicAuth(gin.Accounts{
			"thuanpt": "123@123a",
		}), func(ctx *gin.Context) {
			jsonData, err := ioutil.ReadAll(ctx.Request.Body)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
				return
			}
			type data struct {
				Vmess string `json:"vmess"`
			}
			d := &data{}
			if err := json.Unmarshal(jsonData, &d); err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid json struct"})
				return
			}

			if err := genClashConfig(d.Vmess, *clashCfgTemplatePath, *cOutput); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Error Generate file"})
				return
			}
			ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
		})
		r.Run(serverPort)
	}
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}
	clbContent := clipboard.Read(clipboard.FmtText)
	strContent := *rInput
	if len(strContent) <= 0 {
		strContent = string(clbContent)
	}

}
func genVmessConfig(input string, templatePath, storePath string) error {
	cfg, err := ParseConfigV2ray(input)
	if err != nil {
		return err
	}
	tmpl, err := template.New("vmessConfig").Parse(tmplString)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(storePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	if err := tmpl.Execute(f, cfg); err != nil {
		log.Println(err)
	}
	f.Close()
	return nil
}

func genClashConfig(input string, templatePath, storePath string) error {
	tmplC, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}
	fc, err := os.OpenFile(storePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	cparse, err := ParseConfigClash(input)
	if err != nil {
		return err
	}
	jsonConfig, err := json.Marshal(cparse)
	if err != nil {
		return err
	}
	if err := tmplC.Execute(fc, string(jsonConfig)); err != nil {
		log.Println(err)
	}
	defer fc.Close()
	return nil
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
