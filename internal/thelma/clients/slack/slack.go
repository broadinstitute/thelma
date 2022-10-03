package slackapi

import (
	"github.com/slack-go/slack"
	"os"
	"strings"
	"time"
)

var channelIds = strings.Split(os.Getenv("SLACK_CHANNELS"), ",")

type SlackAPI struct {
	client slack.Client
	token  string
}

func New(token string) (*SlackAPI, error) {
	return &SlackAPI{
		c:     slack.New(token, slack.OptionDebug(true)),
		token: token,
	}, nil
}

func (s SlackAPI) SendDMMessage(owner string) error {
	usr, err := s.client.GetUserByEmail(owner)
	//add error
	bot, err := s.client.GetBotInfo("beebot")
	//add error

	attachment := slack.Attachment{
		Pretext: "Your Bee has finished creating",
		Text:    "Here is the link to your Bee in ArgoCD",
		// Color Styles the Text, making it possible to have like Warnings etc.
		Color: "#36a64f",
		// Fields are Optional extra data!
		Fields: []slack.AttachmentField{
			{
				Title: "Link",
				Value: time.Now().String(),
			},
		},
	}
	// PostMessage will send the message away.
	// First parameter is just the user ID, makes no sense to accept it
	_, timestamp, err := s.client.PostMessage(
		usr.ID,
		slack.MsgOptionAttachments(attachment),
		slack.MsgOptionUser(bot.ID),
	)

	return nil
}

// SimplePostMessage - Fallback, logic-less message post
func SimplePostMessage(c *slack.Client, channelId string, content string) error {
	_, _, err := c.PostMessage(channelId, slack.MsgOptionText(content, false))
	return err
}
