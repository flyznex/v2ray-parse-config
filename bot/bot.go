package bot

import (
	"fmt"
	"os"
	"strconv"
	"v2rayconfig/config"
	"v2rayconfig/utils"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
)

type Bot struct {
	BotApi *tgbotapi.BotAPI
	Conv   map[string]string
}

func InitBot() *Bot {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Error().Err(err).Msg("")
		return nil
	}
	return &Bot{
		BotApi: bot,
		Conv:   make(map[string]string),
	}
}

func (b *Bot) Run(cfg *config.Config, parser ...utils.Parser) {
	subLogger := log.With().Str("mode", "bot").Logger()
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.BotApi.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}
		userId := update.SentFrom().ID
		strUserId := strconv.FormatInt(userId, 10)
		// allow only user in list
		allowed := false

		if len(cfg.TeleBotAllowUsers) == 0 {
			allowed = true
		}

		for _, u := range cfg.TeleBotAllowUsers {
			allowed = u == userId
			if allowed {
				break
			}
		}
		conv, ok := b.Conv[strUserId]
		if !ok {
			b.Conv[strUserId] = ""
		}
		subLogger.Info().Str("Conv", conv).Msg("")
		// Create a new MessageConfig. We don't have text yet,
		// so we leave it empty.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// if update.Message.Text
		// Extract the command from the Message.

		if len(conv) > 0 {
			subLogger.Info().Str("in_conversation", conv).Msg("")
			if update.Message.Command() == "end" {
				msg.Text = "Goodbye"
				b.Conv[strUserId] = ""
			} else {
				if conv == "update" {
					for _, p := range parser {
						if err := p.GenConfig(update.Message.Text); err != nil {
							subLogger.Error().Err(err).Msg("")
							msg.Text = fmt.Sprintf("ERROR: %s. Try again! Or /end to quit.", err.Error())
						} else {
							msg.Text = "Successfull!"
							b.Conv[strUserId] = ""
						}
					}

				}
			}

		} else {
			switch update.Message.Command() {
			case "help":
				msg.Text = "I understand /update and /status"
			case "update":
				msg.Text = "Please insert vmess string to update config file"
				b.Conv[strUserId] = "update"
			case "status":
				msg.Text = "I'm ok"
			default:
				msg.Text = "I don't know that command"
			}
		}

		if _, err := b.BotApi.Send(msg); err != nil {
			log.Error().Err(err).Msg("")
		}
	}

}
