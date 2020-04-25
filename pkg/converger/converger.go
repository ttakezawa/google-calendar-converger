package converger

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	tokenFilePath       = "token.json"
	credentialsFilePath = "credentials.json"
	scope               = calendar.CalendarReadonlyScope
)

func getService() *calendar.Service {
	b, err := ioutil.ReadFile(credentialsFilePath)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, scope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	token := getToken(config)

	ctx := context.Background()
	srv, err := calendar.NewService(
		ctx,
		option.WithTokenSource(config.TokenSource(ctx, token)),
		option.WithScopes(scope),
	)
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}
	return srv
}

func Run() {
	srv := getService()

	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.
		List("primary").
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(t).
		MaxResults(10).
		OrderBy("startTime").
		Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}
	fmt.Println("Upcoming events:")
	if len(events.Items) == 0 {
		fmt.Println("No upcoming events found.")
	} else {
		for _, item := range events.Items {
			date := item.Start.DateTime
			if date == "" {
				date = item.Start.Date
			}
			fmt.Printf("%v (%v)\n", item.Summary, date)
		}
	}
}
