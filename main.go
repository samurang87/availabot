package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/samurang87/availabot/calcheck"
	"github.com/yanzay/tbot"
)

// DefaultHandler receives all messages sent by Telegram to the bot
func DefaultHandler(message *tbot.Message) {
	ctx := context.Background()
	now := time.Now()
	telegramUserID := strconv.FormatInt(int64(message.From.ID), 10)

	if !calcheck.IsAuthenticated(telegramUserID) {
		authURL, err := calcheck.StartAuthFlow(telegramUserID)
		if err != nil {
			log.Println(err)
			message.Replyf("oops: %v", err)
			return
		}
		message.Replyf("@%s auth please: %s", message.From.UserName, authURL)
		return
	}

	busyCal, err := calcheck.GetBusyCalendar(ctx, now, telegramUserID)
	if err != nil {
		log.Println(err)
	}

	result, err := calcheck.GetNextThreeEvenings(now, busyCal)
	if err != nil {
		log.Println(err)
	}

	message.Reply(fmt.Sprint(result))
}

// OAuthHandler handles the OAuth2 redirect URL for Google Auth
func OAuthHandler(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()

	authError := vars.Get("error")
	if authError != "" {
		writeHTTP(w, http.StatusOK, "bummer")
		return
	}

	state := vars.Get("state")
	authCode := vars.Get("code")
	if err := calcheck.CacheGCalToken(r.Context(), state, authCode); err != nil {
		writeHTTP(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeHTTP(w, http.StatusOK, "kthxbai")
}

func writeHTTP(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	if _, err := w.Write([]byte(body)); err != nil {
		log.Println("writeHTTP(): ", err)
	}
}

func main() {
	http.HandleFunc("/oauth2", OAuthHandler)
	go func() {
		log.Fatal(http.ListenAndServe("0.0.0.0:8081", nil))
	}()

	bot, err := tbot.NewServer(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Fatal("unable to start bot server: ", err)
	}

	bot.HandleDefault(DefaultHandler)
	log.Fatal(bot.ListenAndServe())
}
