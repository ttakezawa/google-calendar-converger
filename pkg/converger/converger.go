package converger

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/ttakezawa/google-calendar-converger/pkg/event"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

const (
	tokenFilePath       = "token.json"
	credentialsFilePath = "credentials.json"
	scope               = calendar.CalendarScope

	maxResults = 100

	timeZone     = "Asia/Tokyo"
	calendarId   = "primary"
	colorId      = "11"
	transparency = "opaque"
	visibility   = "private"
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

type Converger struct {
	calendarService *calendar.Service
}

func New() *Converger {
	return &Converger{}
}

func (c *Converger) Init() {
	c.calendarService = getService()
}

func (c *Converger) Run(from time.Time, titlePrefixFilter string, desiredEvents []*event.Event) {
	results, err := c.calendarService.Events.
		List(calendarId).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(from.Format(time.RFC3339)).
		Q(titlePrefixFilter).
		MaxResults(maxResults).
		OrderBy("startTime").
		Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}

	var filteredSourceEvents []*calendar.Event
	for _, e := range results.Items {
		start, err := time.Parse(time.RFC3339, e.Start.DateTime)
		if err != nil {
			log.Fatalf("Unable to parse 'start': %v", err)
		}
		if start.Before(from) {
			// log.Fatalf("skip: %s", start.String())
			continue
		}
		// e.Summary means title.
		if !strings.HasPrefix(e.Summary, titlePrefixFilter) {
			// log.Fatalf("skip: %s", e.Summary)
			continue
		}
		filteredSourceEvents = append(filteredSourceEvents, e)
	}

	var filteredDesiredEvents []*event.Event
	for _, e := range desiredEvents {
		if e.Start.Before(from) {
			continue
		}
		if !strings.HasPrefix(e.Title, titlePrefixFilter) {
			continue
		}
		filteredDesiredEvents = append(filteredDesiredEvents, e)
	}

	sort.SliceStable(filteredDesiredEvents, func(i, j int) bool {
		return filteredDesiredEvents[i].Start.Before(filteredDesiredEvents[j].Start)
	})
	sort.SliceStable(filteredSourceEvents, func(i, j int) bool {
		cmp := strings.Compare(filteredSourceEvents[i].Start.DateTime, filteredSourceEvents[j].Start.DateTime)
		return cmp < 0
	})

	c.deleteUndesiredEvents(filteredDesiredEvents, filteredSourceEvents)
	c.insertDesiredEvents(filteredDesiredEvents, filteredSourceEvents)
}

func (c *Converger) deleteUndesiredEvents(desiredEvents []*event.Event, sourceEvents []*calendar.Event) {
	for _, sourceEvent := range sourceEvents {
		var isDesired bool
		for _, desiredEvent := range desiredEvents {
			if equal(desiredEvent, sourceEvent) {
				isDesired = true
				break
			}
		}
		if isDesired {
			continue
		}
		c.deleteEvent(sourceEvent)
	}
}

func (c *Converger) insertDesiredEvents(desiredEvents []*event.Event, sourceEvents []*calendar.Event) {
	for _, desiredEvent := range desiredEvents {
		var alreadyExists bool
		for _, sourceEvent := range sourceEvents {
			if equal(desiredEvent, sourceEvent) {
				alreadyExists = true
				break
			}
		}
		if alreadyExists {
			continue
		}
		c.insertEvent(desiredEvent)
	}
}

func (c *Converger) deleteEvent(e *calendar.Event) {
	fmt.Printf("delete: %s: %s\n", e.Start.DateTime, e.Summary)
	err := c.calendarService.Events.Delete(calendarId, e.Id).Do()
	if err != nil {
		log.Fatalf("unable to delete event(%s): %v", e.Id, err)
	}
}

func (c *Converger) insertEvent(e *event.Event) {
	fmt.Printf("insert: %s: %s\n", e.Start, e.Title)
	calendarEvent := &calendar.Event{
		Summary:     e.Title,
		Description: e.Description,
		Start: &calendar.EventDateTime{
			DateTime: e.Start.Format(time.RFC3339),
			TimeZone: timeZone,
		},
		End: &calendar.EventDateTime{
			DateTime: e.End.Format(time.RFC3339),
			TimeZone: timeZone,
		},
		ColorId:      colorId,
		Transparency: transparency,
		Visibility:   visibility,
	}
	_, err := c.calendarService.Events.Insert(calendarId, calendarEvent).Do()
	if err != nil {
		log.Fatalf("unable to insert event: %v", err)
	}
}

func equal(e *event.Event, ce *calendar.Event) bool {
	if e.Title != ce.Summary {
		return false
	}
	if e.Description != ce.Description {
		return false
	}
	ceStart, err := time.Parse(time.RFC3339, ce.Start.DateTime)
	if err != nil {
		log.Fatalf("Unable to parse 'start': %v", err)
	}
	ceEnd, err := time.Parse(time.RFC3339, ce.End.DateTime)
	if err != nil {
		log.Fatalf("Unable to parse 'end': %v", err)
	}
	if !e.Start.Equal(ceStart) {

		return false
	}
	if !e.End.Equal(ceEnd) {
		return false
	}
	return true
}
