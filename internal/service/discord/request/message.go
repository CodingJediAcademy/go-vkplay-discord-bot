package request

type Message struct {
	Content    string      `json:"content"`
	Embeds     []Embed     `json:"embeds"`
	Components []Component `json:"components"`
}

type Embed struct {
	Title string `json:"title"`
	Type  string `json:"type"`
	Color string `json:"color"`
	Url   string `json:"url"`
	Image struct {
		Url string `json:"url"`
	} `json:"image"`
	Thumbnail struct {
		Url string `json:"url"`
	} `json:"thumbnail"`
	Author struct {
		Name    string `json:"name"`
		IconUrl string `json:"icon_url"`
	} `json:"author"`
	Fields []Field `json:"fields"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

type Component struct {
	Type       int         `json:"type"`
	Style      int         `json:"style,omitempty"`
	Label      string      `json:"label,omitempty"`
	Url        string      `json:"url,omitempty"`
	Components []Component `json:"components,omitempty"`
}
