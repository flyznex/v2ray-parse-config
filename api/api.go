package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"v2rayconfig/config"
	"v2rayconfig/utils"

	"github.com/gin-gonic/gin"

	"github.com/rs/zerolog"
)

func HomeHandler(cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		updatedAt := utils.ReadUpdatedAt(cfg.ClassCfgOutputPath)
		ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
			"UpdatedAt": updatedAt,
		})
	}
}

func GetClassConfigHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.FileAttachment(cfg.ClassCfgOutputPath, "config.yaml")
	}
}

func GetStatusHandler(cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"updated_at": utils.ReadUpdatedAt(cfg.ClassCfgOutputPath)})
	}
}

func PostUpdateClassConfigHandler(logger zerolog.Logger, cfg *config.Config, classParser utils.Parser) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		jsonData, err := ioutil.ReadAll(ctx.Request.Body)
		if err != nil {
			logger.Error().Err(err).Msg("")
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid input"})
			return
		}
		type data struct {
			Vmess string `json:"vmess"`
		}
		d := &data{}
		if err := json.Unmarshal(jsonData, &d); err != nil {
			logger.Error().Err(err).Msg("")
			ctx.JSON(http.StatusBadRequest, gin.H{"message": "Invalid json struct"})
			return
		}

		if err := classParser.GenConfig(d.Vmess); err != nil {
			logger.Error().Err(err).Str("action", "GenClashConfig").Msg("")
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Error Generate file"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "OK"})
	}
}
