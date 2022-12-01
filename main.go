package main

import (
	"embed"
	"flag"
	"fmt"
	ht "html/template"
	"os"
	"strings"
	"v2rayconfig/api"
	"v2rayconfig/bot"
	"v2rayconfig/config"
	"v2rayconfig/utils"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
	"golang.design/x/clipboard"
)

var (
	serverPort = ":8080"
	//go:embed assets/html/*
	f embed.FS
)

func main() {
	flag.Parse()
	conf := config.InitConfig()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	classAppParser := utils.NewClassAppParser(conf)
	v2rayParser := utils.NewV2rayParser(conf)
	go bot.InitBot().Run(conf, classAppParser)
	if conf.ModeApi {
		runModeApi(conf, classAppParser)
	} else {
		modeCli(conf, classAppParser, v2rayParser)
	}
}

func runModeApi(cfg *config.Config, parser utils.Parser) {
	subLogger := log.With().Str("mode", "api").Logger()
	if p := os.Getenv("PORT"); len(p) > 0 {
		serverPort = fmt.Sprintf(":%s", p)
	}
	subLogger.Info().Str("serverPort", serverPort).Msg("Using api mode")
	acc := os.Getenv("BASIC_AUTH")
	inf := strings.Split(acc, ":")
	if len(inf) != 2 {
		subLogger.Error().Msg("Invalid basic auth config")
		return
	}
	bsauthMid := gin.BasicAuth(gin.Accounts{inf[0]: inf[1]})

	r := gin.Default()
	//load tmpl
	templ := ht.Must(ht.New("").ParseFS(f, "assets/html/*.tmpl"))
	r.SetHTMLTemplate(templ)
	r.GET("/clash-config", api.GetClassConfigHandler(cfg))
	r.GET("/", api.HomeHandler(cfg))
	r.GET("/status", api.GetStatusHandler(cfg))
	r.POST("/config", bsauthMid, api.PostUpdateClassConfigHandler(subLogger, cfg, parser))
	r.Run(serverPort)
}

func modeCli(conf *config.Config, parser ...utils.Parser) {
	subLogger := log.With().Str("mode", "cli").Logger()
	err := clipboard.Init()
	if err != nil {
		subLogger.Err(err)
	}
	strContent := *flag.String("i", "", "Vmess string input")
	if len(strContent) <= 0 && err != nil {
		strContent = string(clipboard.Read(clipboard.FmtText))
	}
	for _, p := range parser {
		if err := p.GenConfig(strContent); err != nil {
			subLogger.Err(err).Str("action", p.GetName())
		}
	}
}
