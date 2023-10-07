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
    <title>gorp</title>

    <style>
        body, body > section, header, form, li {
            display: flex;
        }

        body {
            background-image:
                linear-gradient(rgba(255,255,255,0.5), rgba(255,255,255,0.5)),
                url('data:image/jpeg;base64,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');
            background-size: cover;
            margin-inline: auto;
            width: 60vw;
            min-height: 100vh;
        }

        body > section, ul {
            flex: 1;
        }

        body, body > section {
            flex-direction: column;
        }

        ul {
            list-style: none;
            padding-left: 0;
        }

        header, form {
            justify-content: space-evenly;
        }

        form {
            margin-top: 1rem;
        }

        form > input, button {
            margin-block: auto
        }

        button {
            padding: 0.5rem;
        }

        li {
            gap: 0.5rem;
            justify-content: center;
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
            <input type="text" name="name" placeholder="Name" />
            <textarea name="content" rows="5" cols="40"></textarea>
            <button type="submit">Send</button>
        </form>
    </section>

    <footer>
        <p>Copyright &copy; 2023 Brian Reece
    </footer>
</body>

</html>
`

var l *slog.Logger = slog.Default()

type Message struct {
	From    string
	Content string
}

type Client chan Message

func NewClient() Client {
    l.Info("Creating client")
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
    l.Info("Creating session")
	var s Session = make(map[string]Client)
	return s
}

func (s Session) Broadcast(m Message) {
    l.Info("Broadcasting message")
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
    l.Info("Releasing client")
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

	l.Info("Listening on :3000")
	http.ListenAndServe(":3000", r)
}

func index(w http.ResponseWriter, r *http.Request) {
	l.Info("Rendering page")
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
		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		select {
		case m := <-c.Rx():
			fmt.Fprintf(w, "event: message\ndata: <li><b>%s</b>%s</li>\n\n", m.From, m.Content)
		case <-ctx.Done():
			fmt.Fprint(w, "event: ping\ndata: ping\n\n")
		}

		w.(http.Flusher).Flush()
	}
}

func send(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	s.Broadcast(Message{
		From:    r.FormValue("name"),
		Content: r.FormValue("content"),
	})

	w.WriteHeader(200)
}
