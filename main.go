package main

import (
	"bot/botcore"
	mydb "bot/db"
	"bytes"
	"context"
	"log"
	"os"

	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
	"github.com/robfig/cron"
)

func main() {
	s := mydb.NewStorage("sqlite", "./bot.db")

	bot, bh, err := botcore.CreateBotAndPoll()
	if err != nil {
		log.Fatal(err)
	}

	c := cron.New()

	ChangeAllWeeks := func() {
		ids, err1 := s.Chats.GetAllIds()
		if err1 != nil {
			log.Println(err1)
		}
		for _, id := range ids {
			println(id)
			err2 := botcore.ChangeWeekMain(tu.ID(id), bot, s)
			if err2 != nil {
				//fmt.Print("ОШИБКА")
				log.Println(err2)
			}
		}
	}

	TolstobrowConnection := func() {
		ids, err1 := s.Chats.GetAllIds()
		if err1 != nil {
			log.Println(err1)
		}
		for _, id := range ids {
			println(id)
			chat := s.Chats.Read(id)
			if chat.UseTolstobrow {
				photo, _ := os.ReadFile("./assets/connection.jpg")
				_, err2 := bot.SendPhoto(context.Background(), &telego.SendPhotoParams{MessageThreadID: chat.InfoThread, ParseMode: telego.ModeMarkdownV2, ChatID: telego.ChatID{ID: chat.ID}, Photo: tu.FileFromReader(bytes.NewReader(photo), "connection"), Caption: "[Tolstobrow connection](https://edu.vsu.ru/)"})
				if err2 != nil {
					//fmt.Print("ОШИБКА")
					log.Println(err2)
				}
			}
		}
	}

	c.AddFunc("0 0 0 * * 1", ChangeAllWeeks)
	c.AddFunc("0 30 18 * * 3", TolstobrowConnection)

	//bot.DeleteMyCommands(context.Background(), nil)
	botcore.SetAllCommands(bot)

	botcore.Init(bh, s)
	botcore.ChangeWeekTitle(bh, s)
	botcore.ChangeWeek(bh, s)
	botcore.ChangeTitle(bh, s)
	botcore.SetUsers(bh, s)
	botcore.SetMainThread(bh, s)
	botcore.Ping(bh, s)
	botcore.Tolstobrow(bh, s)
	botcore.AdvertiseWiki(bh, s)
	botcore.AdvertiseTelega(bh, s)

	// Не трогать
	botcore.AddNewPeople(bh, s)
	botcore.DelLeftPeople(bh, s)

	go c.Start()
	defer c.Stop()

	go func() {
		err1 := bh.Start()
		if err1 != nil {
			log.Fatal(err1)
		}
	}()

	select {}
}
