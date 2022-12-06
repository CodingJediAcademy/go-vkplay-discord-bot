package model

import "net/url"

type Guild struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Icon     string    `json:"icon"`
	Channels []Channel `json:"channels"`
}

type Channel struct {
	ID   string `json:"id"`
	Type int    `json:"type"`
}

func (g Guild) IconUrl() string {
	path := "/icons/" + g.ID + "/" + g.Icon + ".png"
	u := url.URL{
		Scheme: "https",
		Host:   "cdn.discordapp.com",
		Path:   path,
	}

	return u.String()
}
