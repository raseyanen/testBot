package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	"log"
	"os"
	"regexp"
	"slices"
	"strings"
)

var adminCmds = []telego.BotCommand{
	{Command: "init", Description: "Проинициализировать бота"},
	{Command: "changeweek", Description: "Вручную сменить ЧИСЛ/ЗНАМ"},
	{Command: "changeweektitle", Description: "Сменить названия ЧИСЛ/ЗНАМ(использовать без [])"},
	{Command: "changetitle", Description: "Сменить название чата(без ЧИСЛ/ЗНАМ)"},
	{Command: "setusers", Description: "Установить список пользователей(для пинга)"},
	{Command: "ping", Description: "Пинг всех(установленных) юзеров(через @)"},
	{Command: "setmainthread", Description: "Установить чат(только для суперчатов) для уведомлений(напр. Толстобров)"},
	{Command: "tolstobrow", Description: "Включить/выключить оповещения на пары Толстоброва"},
}

var userCmds = []telego.BotCommand{
	{Command: "ping", Description: "Пинг всех(установленных) юзеров"},
}

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

	bot, err := telego.NewBot(token, telego.WithDefaultDebugLogger())
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

func ChatInit(bh *th.BotHandler, db *sql.DB) *Chat {
	var chat *Chat
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		if !fromChat(chatID) {
			return nil
		}
		chat = read(chatID.ID, db)
		if chat == nil {
			_, err := ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, ChatID: chatID, Text: "Бот не инициализирован\n/init"})
			if err != nil {
				log.Printf("Ошибка отправки сообщения: %v", err)
			}
			chat = CreateChat(chatID.ID, update.Message.Chat.Title)
			err = write(chat, db)
			if err != nil {
				log.Printf("Ошибка записи чата в БД: %v", err)
			}
		}

		_, err := ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, ChatID: chatID, Text: "Чат успешно инициализирован"})
		if err != nil {
			log.Printf("Ошибка отправки сообщения об успешной инициализации: %v", err)
		}
		return nil
	}, th.CommandEqual("init"))
	return chat
}

func ChangeNumDenum(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		chat := read(chatID.ID, db)
		if chat == nil {
			return nil
		}

		args := strings.Split(update.Message.Text, " ")
		num := args[1]
		denum := args[2]
		chat.Num = num
		chat.Den = denum
		write(chat, db)
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, ParseMode: telego.ModeMarkdownV2, DisableNotification: true, ChatID: chatID, Text: fmt.Sprintf("Успешно.\nЧислитель теперь: `%v`,\nзнаменатель теперь: `%v`", EscapeMarkdown(num), EscapeMarkdown(denum))})
		return nil
	}, th.CommandEqualArgc("changeWeekTitle", 2))
}

func SetUsers(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		chat := read(chatID.ID, db)
		if chat == nil {
			return nil
		}

		people := strings.Split(strings.ReplaceAll(update.Message.Text, ",", ""), " ")[1:]
		flag := true
		for _, peopleStr := range people {
			if peopleStr[0] != '@' {
				flag = false
			}
		}
		if flag {
			chat.Users = []string{}
			chat.Users = append(chat.Users, people...)
			if len(chat.Users) != 0 {
				ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, DisableNotification: true, ChatID: chatID, Text: fmt.Sprintf("Список пользователей\n%v", strings.Join(chat.Users, ","))})
			} else {
				ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, DisableNotification: true, ChatID: chatID, Text: "Список пользователей очищен"})
			}
			write(chat, db)
		} else {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, DisableNotification: true, ChatID: chatID, Text: "Неверный формат команды!"})
		}
		return nil
	}, th.CommandEqual("SetUsers"))
}

func ChangeWeek(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		chat := read(chatID.ID, db)
		if chat == nil {
			return nil
		}
		oldTitle := update.Message.Chat.Title
		numTitle := fmt.Sprintf("[%v] %v", chat.Den, chat.Title)
		denTitle := fmt.Sprintf("[%v] %v", chat.Num, chat.Title)
		if oldTitle != numTitle {
			changeChatTitle(numTitle, chatID, ctx.Bot())
		} else {
			changeChatTitle(denTitle, chatID, ctx.Bot())
		}
		return nil
	}, th.CommandEqual("ChangeWeek"))
}

func ChangeTitle(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		chat := getChatByID(chatID, db, update)
		if chat == nil {
			return nil
		}

		title := strings.Join(strings.Split(update.Message.Text, " ")[1:], " ")
		chat.Title = title
		title = fmt.Sprintf("[%v] %v", chat.Num, title)
		changeChatTitle(title, chatID, ctx.Bot())
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, ChatID: chatID, DisableNotification: true, Text: "Название изменено успешно"})
		write(chat, db)
		return nil
	}, th.CommandEqual("ChangeTitle"))

}

func Ping(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		msgID := update.Message.MessageID
		var pingMsg string
		if len(update.Message.Text) >= 6 {
			pingMsg = update.Message.Text[6:]
		} else {
			pingMsg = update.Message.Text
		}
		chatID := update.Message.Chat.ChatID()
		chat := getChatByID(chatID, db, update)
		if chat == nil {
			return nil
		}

		users := chat.Users
		if len(users) <= 1 {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, DisableNotification: true, ChatID: chatID, Text: "Ошибка: некого пинговать"})
		} else {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: msgID, Quote: pingMsg}, ParseMode: telego.ModeMarkdownV2, ChatID: chatID, Text: "||" + EscapeMarkdown(strings.Join(users, ", ")) + "||"})
		}
		return nil
	}, th.CommandEqual("Ping"))
}

func Tolstobrow(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		if !isAdmin(update.Message.From.ID, ctx.Bot(), chatID) {
			return nil
		}
		chat := getChatByID(chatID, db, update)
		if chat == nil {
			return nil
		}
		var word string
		if chat.UseTolstobrow {
			chat.UseTolstobrow = false
			word = "усыпили"
		} else {
			chat.UseTolstobrow = true
			word = "разбудили"
		}
		write(chat, db)
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, ChatID: chatID, Text: fmt.Sprintf("Вы %v деда", word)})
		return nil
	}, th.CommandEqual("Tolstobrow"))
}

// Пассивные функции \/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\/\
func AddNewPeople(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		chat := getChatByID(chatID, db, update)
		if chat == nil {
			return nil
		}

		if len(update.Message.NewChatMembers) == 0 {
			return nil
		}
		newMembers := update.Message.NewChatMembers
		for _, newMember := range newMembers {
			if !newMember.IsBot {
				chat.Users = append(chat.Users, "@"+newMember.Username)
				write(chat, db)
				ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{DisableNotification: true, ChatID: chatID, Text: "Новый юзер добавлен"})
			}
		}

		return nil
	}, th.AnyMessage())
}

func DelLeftPeople(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		chat := getChatByID(chatID, db, update)
		if chat == nil {
			return nil
		}
		if update.Message.LeftChatMember == nil {
			return nil
		}
		if update.Message.LeftChatMember.IsBot {
			return nil
		}
		username := "@" + update.Message.LeftChatMember.Username
		leftUserIndex := slices.Index(chat.Users, username)
		chat.Users = slices.Delete(chat.Users, leftUserIndex, leftUserIndex+1)
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{DisableNotification: true, ChatID: chatID, Text: "Старый юзер удален"})
		write(chat, db)
		return nil
	}, th.AnyMessage())
}

func AdvertiseWiki(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		bot := ctx.Bot()
		chat := getChatByID(chatID, db, update)
		if chat == nil {
			return nil
		}
		bot.SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, LinkPreviewOptions: &telego.LinkPreviewOptions{IsDisabled: true}, ChatID: chatID, ParseMode: telego.ModeMarkdownV2, DisableNotification: true, Text: "[Ссылка на звездочет](https://star.moodroow.com)"})
		return nil
	}, th.Or(th.TextMatches(regexp.MustCompile(`(?:^|\s|[.,!?])[Зз][Вв][ЕеЁё][Зз][Дд][Оо][Чч][ЕеЁё][Тт](?:\s|[.,!?]|$)`)), th.TextMatches(regexp.MustCompile(`(?:^|\s|[.,!?])[Г][Ии][Тт](?:\s|[.,!?]|$)`)), th.TextMatches(regexp.MustCompile(`(?:^|\s|[.,!?])[Кк][Оо][Сс][Мм][Оо](?:\s|[.,!?]|$)`)), th.TextMatches(regexp.MustCompile(`(?:^|\s|[.,!?])[Вв][Ии][Кк][Ии](?:\s|[.,!?]|$)`))))
}

func AdvertiseTelega(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		bot := ctx.Bot()
		chat := getChatByID(chatID, db, update)
		if chat == nil {
			return nil
		}
		bot.SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, LinkPreviewOptions: &telego.LinkPreviewOptions{IsDisabled: true}, ChatID: chatID, ParseMode: telego.ModeMarkdownV2, DisableNotification: true, Text: "[Ссылка на телегу](https://t.me/starsresearchnews)"})
		return nil
	}, th.TextMatches(regexp.MustCompile(`(?:^|\s|[.,!?])([Тт][Ее][Лл][Ее][Гг]([Аа]|[Ии]|[Ее]|[Уу]|[Oo][Йй]))(?:\s|[.,!?]|$)`)))
}

func SetMainThread(bh *th.BotHandler, db *sql.DB) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		if !isAdmin(update.Message.From.ID, ctx.Bot(), chatID) {
			return nil
		}
		chat := getChatByID(chatID, db, update)
		if chat == nil {
			return nil
		}
		threadID := update.Message.MessageThreadID
		chat.InfoThread = threadID
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, DisableNotification: true, ChatID: chatID, Text: "Тема для уведомлений установлена"})
		write(chat, db)
		return nil
	}, th.CommandEqual("SetMainThread"))
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
