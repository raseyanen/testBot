package botcore

import (
	"bot/db"
	"context"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/mymmrac/telego"
)

func changeChatTitle(title string, chatID telego.ChatID, bot *telego.Bot) {
	ctx := context.Background()
	err := bot.SetChatTitle(ctx, &telego.SetChatTitleParams{ChatID: chatID, Title: title})
	if err != nil {
		bot.SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "У меня нет прав на смену названия данного чата"})
	}
}

func fromChat(id telego.ChatID) bool {
	if id.ID < 0 {
		return true
	}
	return false
}

func isAdmin(userId int64, bot *telego.Bot, id telego.ChatID) bool {
	ctx := context.Background()
	idList := []int64{}
	admins, _ := bot.GetChatAdministrators(ctx, &telego.GetChatAdministratorsParams{ChatID: id})
	for _, admin := range admins {
		idList = append(idList, admin.MemberUser().ID)
	}
	if slices.Contains(idList, userId) {
		return true
	}
	bot.SendMessage(ctx, &telego.SendMessageParams{ChatID: id, Text: "Ошибка! У вас недостаточно прав для выполнения этой команды"})
	return false
}

func getChatByID(chatID telego.ChatID, s *db.Storage, update telego.Update) *db.Chat {
	if !fromChat(chatID) {
		return nil
	}
	chat := s.Chats.Read(chatID.ID)
	if chat == nil {
		//bot.SendMessage(ctx, &telego.SendMessageParams{ChatID: chatID, Text: "Сначала проинициализируйте чат!"})
		chat = db.NewChat(chatID.ID, update.Message.Chat.Title)
		err := s.Chats.Write(chat)
		if err != nil {
			log.Printf("Ошибка записи чата в БД: %v", err)
		}
	}
	return chat
}

func ChangeWeekMain(chatID telego.ChatID, bot *telego.Bot, s *db.Storage) error {
	ctx := context.Background()
	chat := s.Chats.Read(chatID.ID)
	if chat == nil {
		return fmt.Errorf("Нет такого чата")
	}

	// Получаем текущее название чата
	chatInfo, err := bot.GetChat(ctx, &telego.GetChatParams{ChatID: chatID})
	if err != nil {
		return err
	}
	oldTitle := chatInfo.Title

	numTitle := fmt.Sprintf("[%v] %v", chat.Den, chat.Title)
	denTitle := fmt.Sprintf("[%v] %v", chat.Num, chat.Title)

	// Определяем какое название установить
	newTitle := denTitle
	if oldTitle == numTitle {
		newTitle = denTitle
	} else {
		newTitle = numTitle
	}

	// Меняем название
	err = bot.SetChatTitle(ctx, &telego.SetChatTitleParams{
		ChatID: chatID,
		Title:  newTitle,
	})
	return err
}

// EscapeMarkdown экранирует специальные символы Markdown, чтобы они отображались как текст.
func EscapeMarkdown(text string) string {
	// Список символов Markdown, которые нужно экранировать
	specialChars := []string{
		`\`, `*`, `_`, `#`, `>`, `[`, `]`, `(`, `)`, "`", `~`, `-`, `+`, `.`, `!`, `|`,
	}

	result := text
	for _, char := range specialChars {
		// Заменяем каждый специальный символ на экранированную версию
		escapedChar := `\` + char
		result = strings.ReplaceAll(result, char, escapedChar)
	}

	return result
}
