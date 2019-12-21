package calcheck

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

type userInfo struct {
	TelegramUserID   string `json:"tuid"`
	telegramUsername string
	CSRFToken        string `json:"csrf"`
	gcalToken        *oauth2.Token
}

// maps telegram user ID to a userInfo object
var tokenCache = make(map[string]userInfo)
var cacheLock sync.RWMutex

var config = oauth2.Config{
	ClientID:     os.Getenv("BOT_CLIENT_ID"),
	ClientSecret: os.Getenv("BOT_CLIENT_SECRET"),
	Endpoint:     google.Endpoint,
	RedirectURL:  "http://localhost:8081/oauth2",
	Scopes:       []string{calendar.CalendarReadonlyScope},
}

// StartAuthFlow returns an oauth URL with a CSRF token and the Telegram user ID
func StartAuthFlow(telegramUserID string) (string, error) {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	info := userInfo{
		TelegramUserID: telegramUserID,
		CSRFToken:      generateCSRFToken(),
	}
	tokenCache[telegramUserID] = info

	state, err := encodeOAuthState(info)
	if err != nil {
		return "", err
	}

	return config.AuthCodeURL(
		state,
		oauth2.AccessTypeOnline,
		oauth2.SetAuthURLParam("client_id", config.ClientID),
	), nil
}

// IsAuthenticated checks whether a valid oauth token is available for the given Telegram user ID
func IsAuthenticated(telegramUserID string) bool {
	cacheLock.RLock()
	defer cacheLock.RUnlock()

	info, exists := tokenCache[telegramUserID]
	return exists && info.gcalToken != nil && info.gcalToken.Valid()
}

// CacheGCalToken exchanges an auth code for an access token and stores it together with the Telegram user ID
func CacheGCalToken(ctx context.Context, state, gcalAuthCode string) error {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	stateInfo, err := decodeOAuthState(state)
	if err != nil {
		return err
	}

	cachedInfo, exists := tokenCache[stateInfo.TelegramUserID]
	if !exists {
		return fmt.Errorf("no auth flow found for user %s", stateInfo.TelegramUserID)
	}
	if cachedInfo.CSRFToken != stateInfo.CSRFToken {
		return fmt.Errorf("invalid CSRF token")
	}

	gcalToken, err := config.Exchange(ctx, gcalAuthCode)
	if err != nil {
		return err
	}

	cachedInfo.gcalToken = gcalToken
	tokenCache[cachedInfo.TelegramUserID] = cachedInfo
	return nil
}

// GetBusyCalendar retrieves a list of busy slots in the next seven days starting at t0.
func GetBusyCalendar(ctx context.Context, t0 time.Time, telegramUserID string) (cal []*calendar.TimePeriod, err error) {
	userInfo, exists := tokenCache[telegramUserID]
	if !exists {
		return nil, fmt.Errorf("no GCal token for user %s", telegramUserID)
	}

	client := config.Client(ctx, userInfo.gcalToken)
	srv, err := calendar.New(client)
	if err != nil {
		log.Printf("Unable to retrieve calendar Client %v", err)
		return nil, err
	}

	t := t0.Format(time.RFC3339)

	t1 := t0.Add(time.Duration(168) * time.Hour).Format(time.RFC3339)

	fbri := calendar.FreeBusyRequestItem{Id: "primary"}

	query := &calendar.FreeBusyRequest{
		CalendarExpansionMax: 2,
		Items:                []*calendar.FreeBusyRequestItem{&fbri},
		TimeMin:              t,
		TimeMax:              t1,
	}

	freebusy, err := srv.Freebusy.Query(query).Do()
	if err != nil {
		log.Println("Error in executing query to get freebusy calendar")
		return nil, err
	}

	freebusyCal := freebusy.Calendars["primary"].Busy

	return freebusyCal, nil

}

// GetNextThreeEvenings gets a freebusy calendar and a timezone and return the next three free evenings
func GetNextThreeEvenings(t time.Time, c []*calendar.TimePeriod) (free []time.Time, err error) {

	var startDate time.Time

	if t.Hour() < 19 {
		startDate = time.Date(t.Year(), t.Month(), t.Day(), 19, 0, 0, 0, t.Location())
	} else {
		startDate = time.Date(t.Year(), t.Month(), t.Day()+1, 19, 0, 0, 0, t.Location())
	}

	var nextSevenDays [7]time.Time

	for day := 0; day <= 6; day++ {

		if day == 0 {
			nextSevenDays[day] = startDate
		} else {
			nextSevenDays[day] = nextSevenDays[day-1].Add(time.Duration(24) * time.Hour)
		}
	}

	for _, eveningStart := range nextSevenDays {

		eveningEnd := time.Date(
			eveningStart.Year(),
			eveningStart.Month(),
			eveningStart.Day()+1,
			0,
			0,
			0,
			0,
			eveningStart.Location())

		isFree := true

		for _, busySlot := range c {

			startTime, err := time.Parse(time.RFC3339, busySlot.Start)
			if err != nil {
				log.Printf("Unable to parse start time, got this error: %v", err)
				return nil, err
			}

			endTime, err := time.Parse(time.RFC3339, busySlot.End)
			if err != nil {
				log.Printf("Unable to parse end time, got this error: %v", err)
				return nil, err
			}

			if !(endTime.Before(eveningStart) || startTime.After(eveningEnd)) {
				isFree = false
				break
			}
		}

		if isFree {
			free = append(free, eveningStart)
		}

		if len(free) == 3 {
			break
		}
	}
	return
}

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const tokenLen = 32

func generateCSRFToken() string {
	p := make([]byte, tokenLen)

	for i := 0; i < tokenLen; i++ {
		p[i] = chars[rand.Intn(tokenLen)]
	}

	rand.Shuffle(tokenLen, func(i, j int) {
		p[i], p[j] = p[j], p[i]
	})

	return string(p)
}

func encodeOAuthState(info userInfo) (string, error) {
	b, err := json.Marshal(info)
	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}

func decodeOAuthState(state string) (userInfo, error) {
	var info userInfo

	b, err := base64.URLEncoding.DecodeString(state)
	if err != nil {
		return info, err
	}

	err = json.Unmarshal(b, &info)
	return info, err
}
