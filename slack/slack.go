package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type slackField struct {
	Title string `json:"title,omitempty"`
	Value string `json:"value,omitempty"`
	Short bool   `json:"short,omitempty"`
}

type slackAction struct {
	Type  string `json:"type,omitempty"`
	Text  string `json:"text,omitempty"`
	Url   string `json:"url,omitempty"`
	Style string `json:"style,omitempty"`
}

type slackAttachment struct {
	Fallback     string         `json:"fallback,omitempty"`
	Color        string         `json:"color,omitempty"`
	PreText      string         `json:"pretext,omitempty"`
	AuthorName   string         `json:"author_name,omitempty"`
	AuthorLink   string         `json:"author_link,omitempty"`
	AuthorIcon   string         `json:"author_icon,omitempty"`
	Title        string         `json:"title,omitempty"`
	TitleLink    string         `json:"title_link,omitempty"`
	Text         string         `json:"text,omitempty"`
	ImageUrl     string         `json:"image_url,omitempty"`
	Fields       []*slackField  `json:"fields,omitempty"`
	Footer       string         `json:"footer,omitempty"`
	FooterIcon   string         `json:"footer_icon,omitempty"`
	Timestamp    int64          `json:"ts,omitempty"`
	MarkdownIn   []string       `json:"mrkdwn_in,omitempty"`
	Actions      []*slackAction `json:"actions,omitempty"`
	CallbackID   string         `json:"callback_id,omitempty"`
	ThumbnailUrl string         `json:"thumb_url,omitempty"`
}

type slackPayload struct {
	Parse       string             `json:"parse,omitempty"`
	Username    string             `json:"username,omitempty"`
	IconUrl     string             `json:"icon_url,omitempty"`
	IconEmoji   string             `json:"icon_emoji,omitempty"`
	Channel     string             `json:"channel,omitempty"`
	Text        string             `json:"text,omitempty"`
	LinkNames   string             `json:"link_names,omitempty"`
	Attachments []*slackAttachment `json:"attachments,omitempty"`
	UnfurlLinks bool               `json:"unfurl_links,omitempty"`
	UnfurlMedia bool               `json:"unfurl_media,omitempty"`
	Markdown    bool               `json:"mrkdwn,omitempty"`
}

var slackClient = &http.Client{
	Timeout: 15 * time.Second,
}

func sendMessage(url string, payload *slackPayload) error {
	// TODO consider switching to faster json lib
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := slackClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		r, _ := ioutil.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("slack: %s (http status %v)", r, resp.StatusCode)
	}

	return nil
}
