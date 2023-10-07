FROM golang:1.21-alpine
WORKDIR /usr/src/gorp

COPY go.mod go.sum ./
RUN go mod download && \
    go mod verify

COPY . .
RUN go build -v -o /usr/bin/gorp .

CMD [ "/usr/bin/gorp" ]
