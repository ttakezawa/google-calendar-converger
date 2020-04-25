package event

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type Event struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Start       time.Time `json:"start"`
	End         time.Time `json:"end"`
}

func (e *Event) Valid() error {
	var errs []string
	if e.Title == "" {
		errs = append(errs, "'title' should be specified")
	}
	if e.Start.IsZero() {
		errs = append(errs, "'start' should be specified")
	}
	if e.End.IsZero() {
		errs = append(errs, "'end' should be specified")
	}
	if errs == nil {
		return nil
	}
	return errors.New(strings.Join(errs, ","))
}

func Read(r io.Reader) ([]*Event, error) {
	var events []*Event
	if err := json.NewDecoder(os.Stdin).Decode(&events); err != nil {
		return nil, err
	}
	var errs []string
	for i, e := range events {
		if err := e.Valid(); err != nil {
			errs = append(errs, fmt.Sprintf("event %d: %v\n", i, err))
		}
	}
	if errs != nil {
		err := errors.New(strings.Join(errs, ","))
		return nil, err
	}
	return events, nil
}
