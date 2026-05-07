# Groupie Tracker

A web application built entirely in **Go** (standard library only) that consumes the [Groupie Trackers API](https://groupietrackers.herokuapp.com/api) to display information about bands and artists, including their members, concert locations, and dates.

## Features

- **Browse 52+ artists/bands** with album art displayed in a responsive card grid
- **Artist detail pages** with member lists, concert locations, and dates
- **Live search** — type to filter artists in real-time (client-server communication)
- **Multiple data visualizations**: Cards, Tables, Lists, Blocks, Pages
- **Full error handling** — styled error pages for 400, 404, 405, 500
- **Responsive design** — works on desktop, tablet, and mobile
- **Zero external Go dependencies** — uses only the Go standard library

## Tech Stack

| Layer | Technology |
|-------|------------|
| Backend | Go (standard library: `net/http`, `html/template`, `encoding/json`) |
| Frontend | HTML5, CSS3, Vanilla JavaScript |
| API | [Groupie Trackers REST API](https://groupietrackers.herokuapp.com/api) |
| Deployment | Docker (optional) |

## Project Structure

```
groupie-tracker/
├── main.go                  # Entry point: server setup and routes
├── go.mod                   # Go module definition
├── Dockerfile               # Multi-stage Docker build
├── README.md                # This file
│
├── models/
│   └── models.go            # Data structs (Artist, Relation, etc.)
│
├── api/
│   ├── fetch.go             # API fetching with caching and timeout
│   └── fetch_test.go        # Unit tests for API layer
│
├── handlers/
│   ├── handlers.go          # HTTP handlers (Home, Artist, Search, Error)
│   └── handlers_test.go     # Unit tests for handlers
│
├── templates/
│   ├── index.html           # Home page (search block + cards grid)
│   ├── artist.html          # Artist detail (card + list + table)
│   └── error.html           # Error page (400/404/405/500)
│
└── static/
    ├── css/
    │   └── style.css        # Dark theme, responsive, performant animations
    └── js/
        └── script.js        # Live search (client-server event)
```

## How to Run

### Prerequisites

- [Go 1.22+](https://golang.org/dl/) installed

### Run the Server

```bash
go run .
```

The server starts at **http://localhost:8080**

### Build and Run Binary

```bash
go build -o groupie-tracker.exe .
./groupie-tracker.exe
```

## How to Test

Run all unit tests:

```bash
go test ./... -v
```

## Docker

Build and run with Docker:

```bash
docker build -t groupie-tracker .
docker run -p 8080:8080 groupie-tracker
```

## API Endpoints

### Pages (HTML)

| Route | Description |
|-------|-------------|
| `GET /` | Home page with all artist cards |
| `GET /artist?id={id}` | Artist detail page |

### API (JSON)

| Route | Description |
|-------|-------------|
| `GET /search?q={query}` | Search artists by name (returns JSON) |

### Error Handling

| Status | Trigger |
|--------|---------|
| 400 | Missing or invalid artist ID |
| 404 | Unknown page or artist not found |
| 405 | Wrong HTTP method |
| 500 | API failure or internal error |

## Data Visualizations

| Type | Where Used |
|------|------------|
| **Cards** | Home page — artist grid |
| **Blocks** | Search bar section |
| **Tables** | Artist page — concert locations & dates |
| **Lists** | Artist page — band members |
| **Pages** | Each artist has its own detail page |

## External API Used

- **Artists**: `https://groupietrackers.herokuapp.com/api/artists`
- **Locations**: `https://groupietrackers.herokuapp.com/api/locations`
- **Dates**: `https://groupietrackers.herokuapp.com/api/dates`
- **Relations**: `https://groupietrackers.herokuapp.com/api/relation`

## Author

Abdallrhman khalid, Abdullah Khaled, yasser ahmed zaki and Amr Younes
