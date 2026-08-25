package botcore

import "github.com/mymmrac/telego"

var (
	InitCommand            = telego.BotCommand{Command: "init", Description: "Проинициализировать бота"}
	ChangeWeekCommand      = telego.BotCommand{Command: "changeweek", Description: "Вручную сменить ЧИСЛ/ЗНАМ"}
	ChangeWeekTitleCommand = telego.BotCommand{Command: "changeweektitle", Description: "Сменить названия ЧИСЛ/ЗНАМ(использовать без [])"}
	ChangeTitleCommand     = telego.BotCommand{Command: "changetitle", Description: "Сменить название чата(без ЧИСЛ/ЗНАМ)"}
	SetUsersCommand        = telego.BotCommand{Command: "setusers", Description: "Установить список пользователей(для пинга)"}
	PingCommand            = telego.BotCommand{Command: "ping", Description: "Пинг всех(установленных) юзеров(через @)"}
	SetMainThreadCommand   = telego.BotCommand{Command: "setmainthread", Description: "Установить чат(только для суперчатов) для уведомлений(напр. Толстобров)"}
	TolstobrowCommand      = telego.BotCommand{Command: "tolstobrow", Description: "Включить/выключить оповещения на пары Толстоброва"}
)

var adminCmds = []telego.BotCommand{
	InitCommand,
	ChangeWeekCommand,
	ChangeWeekTitleCommand,
	ChangeTitleCommand,
	SetUsersCommand,
	PingCommand,
	SetMainThreadCommand,
	TolstobrowCommand,
}

var userCmds = []telego.BotCommand{
	PingCommand,
}
