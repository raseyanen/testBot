package botcore

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func CreateBotAndPoll() (*telego.Bot, *th.BotHandler, error) {
	// Пытаемся загрузить .env, но не падаем если его нет
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
		// НЕ используем log.Fatal()
	}

	// Проверяем наличие токена
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN environment variable is not set")
	}

	customApiServer := os.Getenv("TELEGRAM_API_SERVER")

	if customApiServer == "" {
		customApiServer = "https://api.telegram.org"
	}

	bot, err := telego.NewBot(token,
		telego.WithDefaultDebugLogger(),
		telego.WithAPIServer(customApiServer))

	if err != nil {
		log.Fatal(err)
		return nil, nil, err
	}

	upd, err := bot.UpdatesViaLongPolling(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
		return nil, nil, err
	}

	bh, _ := th.NewBotHandler(bot, upd)
	return bot, bh, nil
}

func SetAllCommands(bot *telego.Bot) {
	err := bot.SetMyCommands(context.Background(), &telego.SetMyCommandsParams{Commands: adminCmds, Scope: telego.BotCommandScope(&telego.BotCommandScopeAllChatAdministrators{"all_chat_administrators"})})
	if err != nil {
		log.Println(err)
	}
	err = bot.SetMyCommands(context.Background(), &telego.SetMyCommandsParams{Commands: userCmds, Scope: &telego.BotCommandScopeAllGroupChats{"all_group_chats"}})
	if err != nil {
		log.Println(err)
	}
}
