package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

const PAGE string = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">

    <script src="https://unpkg.com/htmx.org@1.9.6" integrity="sha384-FhXw7b6AlE/jyjlZH5iHa/tTe9EpJ1Y55RjcgPbjeWMskSxZt1v9qkxLJWNJaGni" crossorigin="anonymous"></script>
    <script src="https://unpkg.com/htmx.org/dist/ext/sse.js"></script>
    <title>Document</title>

    <style>
        body {
            display: flex;
            flex-direction: column;
            max-width: 100vw;
            min-height: 100vh;
        }

        header, footer {
            flex-grow: 0;
        }

        body > section {
            flex: 1
        }

        ul {
            list-style: none;
        }

        form {
            margin-top: 1rem;
        }

        button {
            margin-top: 0.5rem;
        }

        li {
            display: flex;
            gap: 1rem;
        }
    </style>
</head>

<body>
    <header>
        <h1><a href="/">gorp</a></h1>
    </header>

    <section>
        <ul
            hx-ext="sse"
            sse-connect="/sse"
            sse-swap="message"
            hx-swap="beforeend"
        >
        </ul>

        <form hx-post="/send" hx-swap="none">
            <textarea name="content"></textarea>
            <button type="submit">Send</button>
        </form>
    </section>

    <footer>
        <p>Copyright &copy; 2023 Brian Reece
    </footer>
</body>

</html>
`

type Message struct {
	From    string
	Content string
}

type Client chan Message

func NewClient() Client {
	var c Client = make(chan Message)
	return c
}

func (c Client) Tx() chan<- Message {
	return c
}

func (c Client) Rx() <-chan Message {
	return c
}

type Session map[string]Client

func NewSession() Session {
	var s Session = make(map[string]Client)
	return s
}

func (s Session) Broadcast(m Message) {
	for _, c := range s {
		c.Tx() <- m
	}
}

func (s Session) Client(name string) Client {
	if c, ok := s[name]; ok {
		return c
	} else {
		c = NewClient()
		s[name] = c
		return c
	}
}

func (s Session) Release(name string) {
	delete(s, name)
}

var s Session

func main() {
    s = NewSession()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	r.Get("/", index)
	r.Get("/sse", sse)
	r.Post("/send", send)

	http.ListenAndServe(":3000", r)
}

func index(w http.ResponseWriter, r *http.Request) {
	render.HTML(w, r, PAGE)
}

func sse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Content-Type", "text/event-stream")

	id := middleware.GetReqID(r.Context())
	c := s.Client(id)
	defer s.Release(id)

	for {
        ctx, cancel := context.WithTimeout(context.Background(), 1 * time.Second)
        defer cancel()

        select {
        case m := <-c.Rx():
		    fmt.Fprintf(w, "event: message\ndata: <li><b>%s</b>%s</li>\n\n", m.From, m.Content)
        case <-ctx.Done():
            fmt.Fprintf(w, "event: ping\ndata: ping\n\n")
        }

	w.(http.Flusher).Flush()
	}
}

func send(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	content := r.FormValue("content")

	s.Broadcast(Message{
		From:    middleware.GetReqID(r.Context()),
		Content: content,
	})

	w.WriteHeader(200)
}
