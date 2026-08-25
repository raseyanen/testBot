package botcore

import (
	"bot/db"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strings"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
)

func Init(bh *th.BotHandler, s *db.Storage) *db.Chat {
	var chat *db.Chat
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		if !fromChat(chatID) {
			return nil
		}
		chat = s.Chats.Read(chatID.ID)
		if chat == nil {
			_, err := ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, ChatID: chatID, Text: "Бот не инициализирован\n/init"})
			if err != nil {
				log.Printf("Ошибка отправки сообщения: %v", err)
			}
			chat = db.NewChat(chatID.ID, update.Message.Chat.Title)
			err = s.Chats.Write(chat)
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

func ChangeWeekTitle(bh *th.BotHandler, s *db.Storage) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		chat := s.Chats.Read(chatID.ID)
		if chat == nil {
			return nil
		}

		args := strings.Split(update.Message.Text, " ")
		num := args[1]
		denum := args[2]
		chat.Num = num
		chat.Den = denum
		s.Chats.Write(chat)
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, ParseMode: telego.ModeMarkdownV2, DisableNotification: true, ChatID: chatID, Text: fmt.Sprintf("Успешно.\nЧислитель теперь: `%v`,\nзнаменатель теперь: `%v`", EscapeMarkdown(num), EscapeMarkdown(denum))})
		return nil
	}, th.CommandEqualArgc(ChangeWeekTitleCommand.Command, 2))
}

func SetUsers(bh *th.BotHandler, s *db.Storage) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		chat := s.Chats.Read(chatID.ID)
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
			s.Chats.Write(chat)
		} else {
			ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, DisableNotification: true, ChatID: chatID, Text: "Неверный формат команды!"})
		}
		return nil
	}, th.CommandEqual(SetUsersCommand.Command))
}

func ChangeWeek(bh *th.BotHandler, s *db.Storage) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		chat := s.Chats.Read(chatID.ID)
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
	}, th.CommandEqual(ChangeWeekCommand.Command))
}

func ChangeTitle(bh *th.BotHandler, s *db.Storage) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		fromID := update.Message.From.ID
		if !isAdmin(fromID, ctx.Bot(), chatID) {
			return nil
		}
		chat := getChatByID(chatID, s, update)
		if chat == nil {
			return nil
		}

		title := strings.Join(strings.Split(update.Message.Text, " ")[1:], " ")
		chat.Title = title
		title = fmt.Sprintf("[%v] %v", chat.Num, title)
		changeChatTitle(title, chatID, ctx.Bot())
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, ChatID: chatID, DisableNotification: true, Text: "Название изменено успешно"})
		s.Chats.Write(chat)
		return nil
	}, th.CommandEqual(ChangeTitleCommand.Command))

}

func Ping(bh *th.BotHandler, s *db.Storage) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		msgID := update.Message.MessageID
		var pingMsg string
		if len(update.Message.Text) >= 6 {
			pingMsg = update.Message.Text[6:]
		} else {
			pingMsg = update.Message.Text
		}
		chatID := update.Message.Chat.ChatID()
		chat := getChatByID(chatID, s, update)
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
	}, th.CommandEqual(PingCommand.Command))
}

func Tolstobrow(bh *th.BotHandler, s *db.Storage) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		if !isAdmin(update.Message.From.ID, ctx.Bot(), chatID) {
			return nil
		}
		chat := getChatByID(chatID, s, update)
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
		s.Chats.Write(chat)
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, ChatID: chatID, Text: fmt.Sprintf("Вы %v деда", word)})
		return nil
	}, th.CommandEqual(TolstobrowCommand.Command))
}

func AddNewPeople(bh *th.BotHandler, s *db.Storage) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		chat := getChatByID(chatID, s, update)
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
				s.Chats.Write(chat)
				ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{DisableNotification: true, ChatID: chatID, Text: "Новый юзер добавлен"})
			}
		}

		return nil
	}, th.AnyMessage())
}

func DelLeftPeople(bh *th.BotHandler, s *db.Storage) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		chat := getChatByID(chatID, s, update)
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
		s.Chats.Write(chat)
		return nil
	}, th.AnyMessage())
}

func AdvertiseWiki(bh *th.BotHandler, s *db.Storage) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		bot := ctx.Bot()
		chat := getChatByID(chatID, s, update)
		if chat == nil {
			return nil
		}
		bot.SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, LinkPreviewOptions: &telego.LinkPreviewOptions{IsDisabled: true}, ChatID: chatID, ParseMode: telego.ModeMarkdownV2, DisableNotification: true, Text: "[Ссылка на звездочет](https://star.moodroow.com)"})
		return nil
	}, th.Or(th.TextMatches(regexp.MustCompile(`(?:^|\s|[.,!?])[Зз][Вв][ЕеЁё][Зз][Дд][Оо][Чч][ЕеЁё][Тт](?:\s|[.,!?]|$)`)), th.TextMatches(regexp.MustCompile(`(?:^|\s|[.,!?])[Г][Ии][Тт](?:\s|[.,!?]|$)`)), th.TextMatches(regexp.MustCompile(`(?:^|\s|[.,!?])[Кк][Оо][Сс][Мм][Оо](?:\s|[.,!?]|$)`)), th.TextMatches(regexp.MustCompile(`(?:^|\s|[.,!?])[Вв][Ии][Кк][Ии](?:\s|[.,!?]|$)`))))
}

func AdvertiseTelega(bh *th.BotHandler, s *db.Storage) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		bot := ctx.Bot()
		chat := getChatByID(chatID, s, update)
		if chat == nil {
			return nil
		}
		bot.SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, LinkPreviewOptions: &telego.LinkPreviewOptions{IsDisabled: true}, ChatID: chatID, ParseMode: telego.ModeMarkdownV2, DisableNotification: true, Text: "[Ссылка на телегу](https://t.me/starsresearchnews)"})
		return nil
	}, th.TextMatches(regexp.MustCompile(`(?:^|\s|[.,!?])([Тт][Ее][Лл][Ее][Гг]([Аа]|[Ии]|[Ее]|[Уу]|[Oo][Йй]))(?:\s|[.,!?]|$)`)))
}

func SetMainThread(bh *th.BotHandler, s *db.Storage) {
	bh.Handle(func(ctx *th.Context, update telego.Update) error {
		chatID := update.Message.Chat.ChatID()
		if !isAdmin(update.Message.From.ID, ctx.Bot(), chatID) {
			return nil
		}
		chat := getChatByID(chatID, s, update)
		if chat == nil {
			return nil
		}
		threadID := update.Message.MessageThreadID
		chat.InfoThread = threadID
		ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{ReplyParameters: &telego.ReplyParameters{MessageID: update.Message.MessageID}, DisableNotification: true, ChatID: chatID, Text: "Тема для уведомлений установлена"})
		s.Chats.Write(chat)
		return nil
	}, th.CommandEqual(SetMainThreadCommand.Command))
}
