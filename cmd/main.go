package main

import (
	"context"
	"go-vkplay-discord-bot/internal/service/discord"
	"go-vkplay-discord-bot/internal/service/vkplay"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	disBotToken := os.Getenv("disBotToken")
	vkPlayWssToken := os.Getenv("vkPlayWssToken")
	vkPlayUserID := os.Getenv("vkPlayUserID")

	discordCtx, discordCancel := context.WithCancel(context.Background())
	defer discordCancel()
	disBot := discord.New(discordCtx, disBotToken)
	defer disBot.Stop()

	disBot.Run()

	vkplayCtx, vkplayCancel := context.WithCancel(context.Background())
	defer vkplayCancel()
	spy := vkplay.New(vkplayCtx, vkPlayWssToken, vkPlayUserID, disBot)
	defer spy.Stop()

	//TODO: testing
	disChannelID := os.Getenv("disChannelID")
	streamUrl := url.URL{Scheme: "https", Host: "vkplay.live", Path: "jedi-knight"}
	err := disBot.SendAnnounce(
		"",
		streamUrl,
		1,
		"https://static-cdn.jtvnw.net/previews-ttv/live_user_codingjediknight-320x180.jpg",
		"703256781169885204",
		disChannelID,
	)
	if err != nil {
		log.Println(err)
	}

	// Waiting for term signal
	sig := <-interrupt

	log.Println("cleanup started with", sig, "signal")
	cleanupStart := time.Now()

	// TODO: close all we need
	discordCancel()
	vkplayCancel()
	//time.Sleep(5 * time.Second)

	cleanupElapsed := time.Since(cleanupStart)
	log.Printf("cleanup completed in %v seconds\n", cleanupElapsed.Seconds())

	os.Exit(1)
}
