package slack

import (
	"regexp"

	eventnotifierv1 "github.com/drhelius/event-notifier-operator/api/v1"
	"github.com/go-logr/logr"
	slackclient "github.com/slack-go/slack"
)

// Notification is the configuration for a Slack specific notification
type Notification struct {
	Name    string
	Token   string
	Channel string
	Regex   string
	Log     logr.Logger
}

// Notifications is a list of current Notifications
var Notifications []Notification

// Manage ensures that a Notification instance is under control
func Manage(cr *eventnotifierv1.SlackNotification, log logr.Logger) {

	var n Notification
	n.Name = cr.Name
	n.Token = cr.Spec.Token
	n.Channel = cr.Spec.Channel
	n.Regex = cr.Spec.Regex
	n.Log = log

	found := false

	for _, i := range Notifications {
		if n.Name == i.Name {
			found = true
			i.Token = n.Token
			i.Channel = n.Channel
			i.Regex = n.Regex
			break
		}
	}

	if !found {
		Notifications = append(Notifications, n)
	}
}

// SendMessage sends a message using all current senders
func SendMessage(msg string) {
	for _, n := range Notifications {
		log := n.Log.WithName("Slack")

		re := regexp.MustCompile(n.Regex)

		if len(re.FindAllString(msg, -1)) > 0 {
			api := slackclient.New(n.Token)

			channelID, timestamp, err := api.PostMessage(
				n.Channel,
				slackclient.MsgOptionText(msg, false),
				slackclient.MsgOptionAsUser(true),
			)

			log.Info("Message successfully sent to channel", "Channel", channelID, "Timestamp", timestamp)

			if err != nil {
				log.Error(err, "Unable to send message to Slack")
			}
		}
	}
}
