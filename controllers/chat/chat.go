package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/tjm/puppet-patching-automation/models"
	"github.com/tjm/puppet-patching-automation/views/eventview"
)

// Chat represents the ability to send notifications to a Chat Webhook.
// https://chat.google.com
type Chat struct {
	WebhookURL string
}

// NewChat creates the Chat controller
func NewChat(room *models.ChatRoom) (c *Chat) {
	c = new(Chat)
	c.WebhookURL = room.WebhookURL
	return
}

// HandleEvent sends notifications when events occur.
func (c *Chat) HandleEvent(event *models.Event) {
	// c.WebhookURL = c.WebhookURL + "&threadKey=" + event.ThreadKey // TODO: FIX THIS! IT IS BAD!
	if msg := eventview.PrepareMsg(event); msg != "" {
		makeRequest(c, msg)
	}
}

// // HandleServerStartup sends notifications when KubeWise starts up.
// func (c *Chat) HandleServerStartup(releases []*release.Release) {
// 	if msg := presenters.PrepareServerStartupMsg(releases); msg != "" {
// 		makeRequest(g, msg)
// 	}
// }

func makeRequest(c *Chat, text string) (responseBody []byte) {
	values := map[string]string{"markdown": text}
	jsonValue, err := json.Marshal(values)
	if err != nil {
		log.Error("Error marshaling message into Json", err)
		return
	}

	// resp, requestErr := http.Post(c.WebhookURL, contentType, bytes.NewBuffer(jsonValue))
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.WebhookURL, bytes.NewBuffer(jsonValue))
	if err != nil {
		// Do NOT log the err. It contains the URL which contains sensitive authentication data.
		// If this is to be logged in future, strip the sensitive data from the URL before logging.
		log.Error("Error creating request Chat")
		return
	}
	req.Header.Add("Content-type", "application/json; charset=UTF-8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("Error making httpClient request to Chat", err)
		return
	}

	responseBody, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Error("Malformed response received from Chat", err)
	}

	err = resp.Body.Close()
	if err != nil {
		log.Warn("Error closing response body", err)
	}
	return
}
