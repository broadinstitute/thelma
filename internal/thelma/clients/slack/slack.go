package slack

import (
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/platform"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	slackapi "github.com/slack-go/slack"
)

const configPrefix = "slack"

// Red hex code for setting red background in messages
const colorRed = "#b20000"

// Green hex code for setting green background in messages
const colorGreen = "#33cc33"

type slackConfig struct {
	// Enabled is a flag for whether to enable sending slack messages from thelma
	Enabled bool `default:"true"`
	// Slack Token is a slack bot token credential expected to be provided to thelma in the envionment
	// NOTE: this replaces previous functionality where Thelma would dynamically reach out to vault
	// to fetch a token at runtime.
	Token      string
	ChannelIDs struct {
		DevopsAlerts string `default:"C011NQS8Q2Z"` // Devops alerts channel: #ap-k8s-monitor
	}
}

type Slack interface {
	// SendDevopsAlert posts a message to the configured DevOps alert channel
	SendDevopsAlert(title string, text string, ok bool) error
	// SendDirectMessage send a direct message to a user, by email
	SendDirectMessage(email string, markdown string) error
}

type slack struct {
	client *slackapi.Client
	cfg    slackConfig
}

// New is lazy, meaning that it does not try to set up its Slack client immediately.
// This makes it much less likely to error, so that a Slack client can be safely passed
// around Thelma and only try-caught at the actual call-site of attempting to send a
// message
func New(thelmaConfig config.Config) (Slack, error) {
	var cfg slackConfig
	if err := thelmaConfig.Unmarshal(configPrefix, &cfg); err != nil {
		return nil, err
	}
	return &slack{
		cfg: cfg,
	}, nil
}

// SendDevopsAlert posts a message to the configured DevOps alert channel
func (s *slack) SendDevopsAlert(title string, text string, ok bool) error {
	if err := s.requireClient(); err != nil {
		return err
	}

	channelId := s.cfg.ChannelIDs.DevopsAlerts

	var color string
	if ok {
		color = colorGreen
	} else {
		color = colorRed
	}

	_, _, err := s.client.PostMessage(channelId, slackapi.MsgOptionAttachments(slackapi.Attachment{
		Color:     color,
		Title:     title,
		TitleLink: platform.Lookup().Link(),
		Text:      text,
	}))

	if err != nil {
		return errors.Errorf("couldn't send message to channel %s: %v", channelId, err)
	}
	return nil
}

// SendDirectMessage send a direct message to a user, by email
func (s *slack) SendDirectMessage(email string, markdown string) error {
	if err := s.requireClient(); err != nil {
		log.Info().Msg("will continue without sending Slack messages; enable debug logging for more information")
		log.Debug().Msgf("failed to initialize Slack client: %v", err)
		return nil
	}
	user, err := s.client.GetUserByEmail(email)
	if err != nil {
		return errors.Errorf("couldn't get user for %s: %v", email, err)
	} else if user == nil {
		return errors.Errorf("couldn't get user for %s: Slack didn't error but didn't return a user object either", email)
	}
	channel, _, _, err := s.client.OpenConversation(&slackapi.OpenConversationParameters{
		Users: []string{user.ID},
	})
	if err != nil {
		return errors.Errorf("couldn't open channel for %s user %s: %v", email, user.ID, err)
	} else if channel == nil {
		return errors.Errorf("couldn't open channel for %s user %s: Slack didn't error but didn't return a channel object either", email, user.ID)
	}
	_, _, err = s.client.PostMessage(channel.ID, slackapi.MsgOptionBlocks(slackapi.SectionBlock{
		Type: slackapi.MBTSection,
		Text: &slackapi.TextBlockObject{
			Type: slackapi.MarkdownType,
			Text: markdown,
		},
	}))
	if err != nil {
		return errors.Errorf("couldn't send message to %s user channel %s: %v", email, channel.ID, err)
	}
	return nil
}

func (s *slack) requireClient() error {
	if s.client == nil {
		if s.cfg.Enabled {
			var token string
			if s.cfg.Enabled && s.cfg.Token != "" {
				token = s.cfg.Token
			}
			s.client = slackapi.New(token)
		}
	}
	if s.client == nil {
		return errors.Errorf("could not build Slack client")
	} else {
		return nil
	}
}
