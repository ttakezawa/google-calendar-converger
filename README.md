# google-calendar-converger

## Usage

### 1. Install

```sh
go get github.com/ttakezawa/google-calendar-converger/cmd/google-calendar-converger
```

### 2. Create your own client credential

- Go to the [Google developer console](https://console.developers.google.com/)
- Make a new project for google-calendar-converger
- Enable the Calendar API
- Create a new Client ID
- Download credentials as "credentials.json"

### 3. Login with OAuth2

```sh
google-calendar-converger -init
```

### 4. sync events

```sh
cat examples/meal-events-desired.json
[
  {
    "title": "meal:breakfast",
    "description": "nice breakfast",
    "start": "2021-04-30T08:15:00+09:00",
    "end": "2021-04-30T09:00:00+09:00"
  },
  {
    "title": "meal:dinner",
    "description": "I'll have a dinner with my friends.",
    "start": "2021-04-30T18:45:00+09:00",
    "end": "2021-04-30T19:30:00+09:00"
  }
]

google-calendar-converger -title-prefix-filter "meal:" < examples/meal-events-desired.json
```
