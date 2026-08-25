package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type Chat struct {
	ID            int64    `bson:"id"`
	InfoThread    int      `bson:"main_topic"`
	Num           string   `bson:"num"`
	Den           string   `bson:"den"`
	Title         string   `bson:"title"`
	Users         []string `bson:"users"`
	UseTolstobrow bool     `bson:"use_tolstobrow"`
}

type ChatRepository struct {
	db *sql.DB
}

func (cr *ChatRepository) Write(chat *Chat) error {
	db := cr.db
	usersJson, err := json.Marshal(chat.Users)
	if err != nil {
		return err
	}
	err = cr.Delete(chat.ID)
	if err != nil {
		return err
	}
	result, exec := db.Exec("INSERT INTO chats (id, main_topic, num, den, title, users, use_tolstobrow) VALUES (?, ?, ?, ?, ?, ?,?)", chat.ID, chat.InfoThread, chat.Num, chat.Den, chat.Title, usersJson, chat.UseTolstobrow)
	if exec != nil {
		return exec
	}
	fmt.Println(result.LastInsertId())
	return nil
}

func (cr *ChatRepository) Read(id int64) *Chat {
	row := cr.db.QueryRow("SELECT id, main_topic, num, den, title, users, use_tolstobrow FROM chats WHERE id = ?", id)

	var (
		dbId          int64
		mainTopic     int
		num           string
		den           string
		title         string
		usersJSON     string
		useTolstobrow bool
	)

	err := row.Scan(&dbId, &mainTopic, &num, &den, &title, &usersJSON, &useTolstobrow)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return nil
	}

	// Декодируем JSON массив пользователей
	var usersSlice []string
	if err := json.Unmarshal([]byte(usersJSON), &usersSlice); err != nil {
		return nil
	}

	// Создаем и возвращаем объект Chat
	chat := Chat{
		ID:            dbId,
		InfoThread:    mainTopic,
		Num:           num,
		Den:           den,
		Title:         title,
		Users:         usersSlice,
		UseTolstobrow: useTolstobrow,
	}

	return &chat
}

func NewChat(id int64, title string) *Chat {
	chat := Chat{}
	chat.ID = id
	chat.InfoThread = 0
	chat.Num = "Числитель"
	chat.Den = "Знаменатель"
	chat.Title = title
	chat.Users = []string{}
	chat.UseTolstobrow = false
	return &chat
}

func (c Chat) ToString() string {
	return fmt.Sprintf("%v %v %v %v %v %v", c.ID, c.InfoThread, c.Num, c.Den, c.Title, c.Users)
}
func (cr *ChatRepository) Delete(id int64) error {
	result, err := cr.db.Exec("DELETE FROM chats WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete chat: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("chat with id %d not found", id)
	}

	return nil
}

func (cr *ChatRepository) GetAllIds() ([]int64, error) {
	db := cr.db
	raw, err := db.Query("SELECT id FROM chats")
	if err != nil {
		return nil, err
	}
	defer raw.Close()

	var ids []int64

	for raw.Next() {
		var id int64

		err = raw.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if err != nil {
		return nil, err
	}
	return ids, nil
}
