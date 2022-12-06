package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"go-vkplay-discord-bot/internal/service/discord/event"
	"go-vkplay-discord-bot/internal/service/discord/model"
	"go-vkplay-discord-bot/internal/service/discord/request"
	"io"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"time"
)

const (
	eventGuildCreate = "GUILD_CREATE"

	opDispatch = 0
	opHello    = 10
)

var (
	wssAddr = url.URL{
		Scheme:   "wss",
		Host:     "gateway.discord.gg",
		RawQuery: "v=10&encoding=json",
	}
	heartbeatInterval = time.Duration(45) * time.Second
)

type Bot struct {
	ctx    context.Context
	conn   *websocket.Conn
	token  string
	Guilds []*model.Guild
}

func New(
	ctx context.Context,
	token string,
) Bot {
	wssConn, _, err := websocket.DefaultDialer.DialContext(
		ctx,
		wssAddr.String(),
		nil,
	)
	if err != nil {
		log.Fatalln("dial:", err)
	}

	bot := Bot{
		ctx:   ctx,
		conn:  wssConn,
		token: token,
	}

	return bot
}

func (b Bot) Stop() {
	err := b.conn.Close()
	if err != nil {
		log.Println(err)
	}
}

func (b Bot) SendAnnounce(
	title string,
	channelUrl url.URL,
	viewers int,
	streamFrameUrl string,
	guildID string,
	channelID string,
) error {
	var currentGuild model.Guild
	for _, guild := range b.Guilds {
		if guild.ID == guildID {
			currentGuild = *guild
		}
	}
	field := request.Field{
		Name:   "Viewers",
		Value:  strconv.Itoa(viewers),
		Inline: false,
	}
	embed := request.Embed{
		Title: title,
		Type:  "rich",
		Color: "008486",
		Url:   channelUrl.String(),
		Image: struct {
			Url string `json:"url"`
		}{
			Url: streamFrameUrl,
		},
		Thumbnail: struct {
			Url string `json:"url"`
		}{
			Url: currentGuild.IconUrl(),
		},
		Author: struct {
			Name    string `json:"name"`
			IconUrl string `json:"icon_url"`
		}{
			Name:    currentGuild.Name,
			IconUrl: currentGuild.IconUrl(),
		},
		Fields: []request.Field{field},
	}
	component := request.Component{
		Type:  2,
		Style: 5,
		Label: "Залетай на стрим!",
		Url:   channelUrl.String(),
	}
	box := request.Component{
		Type:       1,
		Components: []request.Component{component},
	}
	reqJson := request.Message{
		Content:    fmt.Sprintf("Привет @everyone, CodingJediKnight подрубил трансляцию! %s", channelUrl.String()),
		Embeds:     []request.Embed{embed},
		Components: []request.Component{box},
	}

	for _, channel := range currentGuild.Channels {
		if channel.ID == channelID {
			reqUrl := url.URL{
				Scheme: "https",
				Host:   "discord.com",
				Path:   fmt.Sprintf("/api/v10/channels/%s/messages", channelID),
			}
			jsonStr, _ := json.Marshal(reqJson)
			log.Println(jsonStr)
			req, err := http.NewRequest("POST", reqUrl.String(), bytes.NewBuffer(jsonStr))
			req.Header.Set("User-Agent", "DiscordBot (go-vkplay-discord-bot, 0.0.1)")
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", fmt.Sprintf("Bot %s", b.token))

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				panic(err)
			}
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(resp.Body)

			if resp.StatusCode != http.StatusOK {
				return errors.New("failed on sending announce")
			}

			return nil
		}
	}

	return errors.New("smthing went wrong")
}

func (b Bot) Run() {
	done := make(chan struct{})
	messageOut := make(chan []byte)

	go func() {
		defer close(done)
		for {
			_, message, err := b.conn.ReadMessage()
			if err != nil {
				log.Println("[bot] read:", err)
				return
			}

			unmarshalledMsg := event.Payload{}
			err = json.Unmarshal(message, &unmarshalledMsg)
			if err != nil {
				log.Println("[bot] unmarshal:", err)

			}

			//log.Printf("[bot] recv: %s", message)

			authPayload := event.Payload{
				OperationCode: 2,
				Data: map[string]interface{}{
					"token":   b.token,
					"intents": 513,
					"properties": map[string]string{
						"os":      runtime.GOOS,
						"browser": "go-vkplay-discord-bot",
						"device":  "go-vkplay-discord-bot",
					},
				},
			}

			switch unmarshalledMsg.OperationCode {
			case opHello:
				log.Printf("[bot] Send Sub Details")
				authMsg, err := json.Marshal(authPayload)
				if err != nil {
					log.Println("[bot] marshal:", err)
				}
				messageOut <- authMsg
			case opDispatch:
				if unmarshalledMsg.EventName == eventGuildCreate {
					guild := model.Guild{}
					data, _ := json.Marshal(unmarshalledMsg.Data)
					_ = json.Unmarshal(data, &guild)
					b.Guilds = append(b.Guilds, &guild)
				}
			}
		}
	}()

	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	for {
		select {
		case <-done:
			return
		case m := <-messageOut:
			//log.Printf("[bot] Send Message %s", string(m))
			err := b.conn.WriteMessage(websocket.TextMessage, m)
			if err != nil {
				log.Println("[bot] write:", err)
				return
			}
		case <-heartbeat.C:
			log.Println("[bot] heart beating...")
			err := b.conn.WriteMessage(websocket.TextMessage, []byte("{\"op\": 1,\"d\": 251}"))
			if err != nil {
				log.Println("[bot] write:", err)
				return
			}
		case <-b.ctx.Done():
			log.Println("[bot] write interrupt")

			err := b.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("[bot] write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
