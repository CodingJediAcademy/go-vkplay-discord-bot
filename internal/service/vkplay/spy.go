package vkplay

import (
	"context"
	"github.com/gorilla/websocket"
	"go-vkplay-discord-bot/internal/contract"
	"log"
	"net/http"
	"net/url"
)

var (
	wssAddr = url.URL{
		Scheme: "wss",
		Host:   "pubsub.boosty.to",
		Path:   "/connection/websocket",
	}
)

type Spy struct {
	ctx        context.Context
	conn       *websocket.Conn
	announcers []contract.Announcer
	wssToken   string
	userID     string
}

func New(
	ctx context.Context,
	wssToken string,
	userID string,
	announcers ...contract.Announcer,
) Spy {
	headers := http.Header{}
	headers.Add("Origin", "https://vkplay.live")

	wssConn, _, err := websocket.DefaultDialer.DialContext(
		ctx,
		wssAddr.String(),
		headers,
	)
	if err != nil {
		log.Fatalln("dial:", err)
	}

	spy := Spy{
		ctx:        ctx,
		conn:       wssConn,
		announcers: announcers,
		wssToken:   wssToken,
		userID:     userID,
	}

	return spy
}

func (s Spy) Stop() {
	err := s.conn.Close()
	if err != nil {
		log.Println(err)
	}
}
