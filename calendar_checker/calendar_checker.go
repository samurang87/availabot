package calendar_checker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("calendar-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

// GetBusyCalendar retrieves a list of busy slots in the next seven days starting at t0.
func GetBusyCalendar(t0 time.Time) (start time.Time, cal []*calendar.TimePeriod, err error) {

	ctx := context.Background()

	b, err := ioutil.ReadFile("client_id.json")
	if err != nil {
		log.Printf("Unable to read client secret file: %v", err)
		return start, nil, err
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/calendar-go-quickstart.json
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Printf("Unable to parse client secret file to config: %v", err)
		return start, nil, err
	}
	client := getClient(ctx, config)

	srv, err := calendar.New(client)
	if err != nil {
		log.Printf("Unable to retrieve calendar Client %v", err)
		return start, nil, err
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
		return start, nil, err
	}

	freebusyCal := freebusy.Calendars["primary"].Busy

	return start, freebusyCal, nil

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
