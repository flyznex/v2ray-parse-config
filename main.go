package main

import (
	"bufio"
	"embed"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	ht "html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

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
	//go:embed assets/html/*
	f embed.FS
)

func parseVmessString(input string) (map[string]interface{}, error) {
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
	out, err := parseVmessString(input)
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

var (
	rInput               = flag.String("i", "", "Vmess string input")
	v2RayTemplatePath    = flag.String("r", "./assets/template/v2ray-config.tml", "V2ray config template")
	rOuput               = flag.String("f", "./v2ray-config.json", "V2ray Config Json Path")
	clashCfgTemplatePath = flag.String("t", "./assets/template/config.tmpl", "template config")
	cOutput              = flag.String("c", "./config.yaml", "Clash config output path")
	modeApi              = flag.Bool("m", true, "enable mod api")
)

func main() {
	flag.Parse()
	if *modeApi {
		runModeApi()
	} else {
		modeCli()
	}

}

func runModeApi() {
	if p := os.Getenv("PORT"); len(p) > 0 {
		serverPort = fmt.Sprintf(":%s", p)
	}
	acc := os.Getenv("BASIC_AUTH")
	inf := strings.Split(acc, ":")
	if len(inf) != 2 {
		fmt.Println("[MODE: API] [ERROR]: Invalid basic auth config")
		return
	}

	r := gin.Default()
	//load tmpl
	templ := ht.Must(ht.New("").ParseFS(f, "assets/html/*.tmpl"))
	r.SetHTMLTemplate(templ)
	r.GET("/clash-config", func(c *gin.Context) {
		c.FileAttachment(string(*cOutput), "config.yaml")
	})
	r.GET("/", func(ctx *gin.Context) {
		updatedAt := readUpdatedAt()
		ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
			"UpdatedAt": updatedAt,
		})
	})
	r.GET("/status", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"updated_at": readUpdatedAt()})
	})
	r.POST("/config", gin.BasicAuth(gin.Accounts{
		inf[0]: inf[1],
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

func modeCli() {
	err := clipboard.Init()
	if err != nil {
		panic(err)
	}
	clbContent := clipboard.Read(clipboard.FmtText)
	strContent := *rInput
	if len(strContent) <= 0 {
		strContent = string(clbContent)
	}
	if err := genVmessConfig(strContent, *v2RayTemplatePath, *rOuput); err != nil {
		fmt.Println("[MODE: CLI] [GenVMESS] [ERROR]: ", err.Error())
	}
	if err := genClashConfig(strContent, *clashCfgTemplatePath, *cOutput); err != nil {
		fmt.Println("[MODE: CLI] [GenClashConfig] [ERROR]: ", err.Error())
	}
}

func readUpdatedAt() string {
	f, err := os.Open(*cOutput)
	if err != nil {
		fmt.Println("[ERROR] Get updated time got error", err.Error())
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var line int
	for scanner.Scan() {
		if line == 0 {
			return scanner.Text()[19:]
		}
	}
	return ""
}
func genVmessConfig(input string, templatePath, storePath string) error {
	cfg, err := ParseConfigV2ray(input)
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
	type cfg struct {
		Value     string
		UpdatedAt string
	}
	data := cfg{
		Value:     string(jsonConfig),
		UpdatedAt: time.Now().In(time.FixedZone("UTC+7", +7*60*60)).Format("2006-01-02T15:04:05"),
	}
	if err := tmplC.Execute(fc, data); err != nil {
		log.Println(err)
	}
	defer fc.Close()
	return nil
}
