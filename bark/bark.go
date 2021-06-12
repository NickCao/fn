package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"gopkg.in/tucnak/telebot.v2"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var bark chi.Router

func MustLookupEnv(key string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	} else {
		panic(fmt.Sprintf("environment variable %s is required", key))
	}
}

func Serve(fn http.HandlerFunc) {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Mount("/", fn)
	if err := http.ListenAndServe(":8080", router); err != nil {
		panic(err)
	}
}

func init() {
	chatID, err := strconv.ParseInt(MustLookupEnv("CHAT_ID"), 10, 64)
	if err != nil {
		panic(err)
	}
	botToken := MustLookupEnv("BOT_TOKEN")
	webhookPath := MustLookupEnv("WEBHOOK_PATH")
	bot, err := telebot.NewBot(telebot.Settings{
		Token: botToken,
		Poller: telebot.NewMiddlewarePoller(&telebot.LongPoller{Timeout: 10 * time.Second}, func(upd *telebot.Update) bool {
			return false
		}),
	})
	if err != nil {
		panic(err)
	}
	bark = chi.NewRouter()
	bark.Post(webhookPath, func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
		}
		_, err = bot.Send(telebot.ChatID(chatID), string(payload))
		if err != nil {
			log.Println(err)
		}
	})
}

func main() {
	Serve(bark.ServeHTTP)
}
