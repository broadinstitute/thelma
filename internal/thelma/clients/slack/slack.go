package slack

import (
	"fmt"
	"github.com/broadinstitute/thelma/internal/thelma/app/config"
	"github.com/broadinstitute/thelma/internal/thelma/app/logging"
	"github.com/broadinstitute/thelma/internal/thelma/app/platform"
	vaultapi "github.com/hashicorp/vault/api"
	"github.com/rs/zerolog/log"
	"github.com/slack-go/slack"
)

const configPrefix = "slack"

// Red hex code for setting red background in messages
const colorRed = "#b20000"

// Green hex code for setting green background in messages
const colorGreen = "#33cc33"

type slackConfig struct {
	// Vault is the only (current) mechanism to obtain Slack credentials, since we want to always authenticate as
	// a bot user instead of relying on user-credentials when running locally.
	Vault struct {
		Enabled bool   `default:"true"`
		Path    string `default:"secret/suitable/beehive/prod/slack"`
		Key     string `default:"bot-user-oauth-token"`
	}
	ChannelIDs struct {
		DevopsAlerts string `default:"C011NQS8Q2Z"` // Devops alerts channel: #ap-k8s-monitor
	}
}

type Slack struct {
	client             *slack.Client
	cfg                slackConfig
	vaultClientFactory func() (*vaultapi.Client, error)
}

// New is lazy, meaning that it does not try to set up its Vault client immediately.
// This makes it much less likely to error, so that a Slack client can be safely passed
// around Thelma and only try-caught at the actual call-site of attempting to send a
// message
func New(thelmaConfig config.Config, vaultClientFactory func() (*vaultapi.Client, error)) (*Slack, error) {
	var cfg slackConfig
	if err := thelmaConfig.Unmarshal(configPrefix, &cfg); err != nil {
		return nil, err
	}
	return &Slack{
		cfg:                cfg,
		vaultClientFactory: vaultClientFactory,
	}, nil
}

// SendDevopsAlert posts a message to the configured DevOps alert channel
func (s *Slack) SendDevopsAlert(title string, text string, ok bool) error {
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

	_, _, err := s.client.PostMessage(channelId, slack.MsgOptionAttachments(slack.Attachment{
		Color:     color,
		Title:     title,
		TitleLink: platform.Lookup().Link(),
		Text:      text,
	}))

	if err != nil {
		return fmt.Errorf("couldn't send message to channel %s: %v", channelId, err)
	}
	return nil
}

// SendDirectMessage send a direct message to a user, by email
func (s *Slack) SendDirectMessage(email string, markdown string) error {
	if err := s.requireClient(); err != nil {
		return err
	}
	user, err := s.client.GetUserByEmail(email)
	if err != nil {
		return fmt.Errorf("couldn't get user for %s: %v", email, err)
	} else if user == nil {
		return fmt.Errorf("couldn't get user for %s: Slack didn't error but didn't return a user object either", email)
	}
	channel, _, _, err := s.client.OpenConversation(&slack.OpenConversationParameters{
		Users: []string{user.ID},
	})
	if err != nil {
		return fmt.Errorf("couldn't open channel for %s user %s: %v", email, user.ID, err)
	} else if channel == nil {
		return fmt.Errorf("couldn't open channel for %s user %s: Slack didn't error but didn't return a channel object either", email, user.ID)
	}
	_, _, err = s.client.PostMessage(channel.ID, slack.MsgOptionBlocks(slack.SectionBlock{
		Type: slack.MBTSection,
		Text: &slack.TextBlockObject{
			Type: slack.MarkdownType,
			Text: markdown,
		},
	}))
	if err != nil {
		return fmt.Errorf("couldn't send message to %s user channel %s: %v", email, channel.ID, err)
	}
	return nil
}

func (s *Slack) requireClient() error {
	if s.client == nil {
		if s.cfg.Vault.Enabled {
			vaultClient, err := s.vaultClientFactory()
			if err != nil {
				return err
			}
			token, err := readTokenFromVault(s.cfg, vaultClient)
			if err != nil {
				return err
			}
			s.client = slack.New(token)
		}
	}
	if s.client == nil {
		return fmt.Errorf("could not build Slack client")
	} else {
		return nil
	}
}

func readTokenFromVault(cfg slackConfig, vaultClient *vaultapi.Client) (string, error) {
	log.Debug().Msgf("Attempting to read Slack token from Vault (%s)", cfg.Vault.Path)
	secret, err := vaultClient.Logical().Read(cfg.Vault.Path)
	if err != nil {
		return "", fmt.Errorf("error loading Slack token from Vault path %s: %v", cfg.Vault.Path, err)
	}
	v, exists := secret.Data[cfg.Vault.Key]
	if !exists {
		return "", fmt.Errorf("error loading Slack token from Vault path %s: missing key %s", cfg.Vault.Path, cfg.Vault.Key)
	}
	asStr, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("error loading Slack token from Vault path %s: expecting string key value for %s", cfg.Vault.Path, cfg.Vault.Key)
	}
	logging.MaskSecret(asStr)
	return asStr, nil
}
