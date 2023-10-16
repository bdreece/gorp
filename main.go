package main

import (
	"fmt"
	"log/slog"
	"net/http"

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
        body, body > section, header, form, li, .form-control {
            display: flex;
        }

        body {
            font-family: system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif;
            margin-inline: auto;
            width: 60vw;
            min-height: 100vh;
        }

        body > section, ul {
            flex: 1;
        }

        body, body > section, .form-control {
            flex-direction: column;
        }

        ul {
            list-style: none;
            padding-left: 0;
        }

        form {
            margin-top: 1rem;
        }

        @media screen and (max-width: 640px) {
            form {
                flex-direction: column;
            }
        }

        form > input, i {
            margin-block: auto
        }

        button {
            margin-block: 1.5rem auto;
        }

        form, li, header {
            gap: 1rem;
        }

        button, li {
            padding: 0.5rem;
        }
    </style>
</head>

<body>
    <header>
        <h1><a href="/">gorp</a></h1>

        <i>At consectetur lorem donec massa sapien faucibus et molestie ac</i>
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
            <div class="form-control">
                <label for="name">Name</label>
                <input type="text" id="name" name="name" placeholder="Name" />
            </div>

            <div class="form-control">
                <label for="content">Message</label>
                <textarea id="content" name="content" rows="5" cols="40"></textarea>
            </div>

            <button type="submit">Send</button>
        </form>
    </section>

    <footer>
        <p>Copyright &copy; 2023 <a href="https://www.bdreece.dev" target="_blank" rel="noreferrer">Brian Reece</a></p>
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
		select {
		case m := <-c.Rx():
			fmt.Fprintf(w, "event: message\ndata: <li><b>%s</b>%s</li>\n\n", m.From, m.Content)
        case <-r.Context().Done():
            goto done
		}

		w.(http.Flusher).Flush()
	}

done:
}

func send(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	s.Broadcast(Message{
		From:    r.FormValue("name"),
		Content: r.FormValue("content"),
	})

	w.WriteHeader(200)
}
