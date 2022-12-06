package contract

import "net/url"

type Announcer interface {
	SendAnnounce(
		title string,
		channelUrl url.URL,
		viewers int,
		streamFrameUrl string,
		guildID string,
		channelID string,
	) error
}
