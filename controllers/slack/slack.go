package slack

import (
	"regexp"
	"strconv"

	eventnotifierv1 "github.com/drhelius/event-notifier-operator/api/v1"
	"github.com/go-logr/logr"
	"github.com/slack-go/slack"
	slackclient "github.com/slack-go/slack"
	corev1 "k8s.io/api/core/v1"
)

// Notification is the configuration for a Slack specific notification item
type Notification struct {
	Name string
	Spec *eventnotifierv1.SlackNotificationSpec
}

// Notifications is a list of current Notification items
var Notifications []Notification

// Manage ensures that a Notification instance is under control
func Manage(cr *eventnotifierv1.SlackNotification) {

	var newNotification Notification
	newNotification.Name = cr.Name
	newNotification.Spec = cr.Spec.DeepCopy()

	found := false

	// for each Notification in our current list
	for i, n := range Notifications {
		// check if we have this Notification already in the list
		if n.Name == newNotification.Name {
			// if we already have this notification in the list we want
			// to update it in case it has changed
			found = true
			Notifications[i] = newNotification
			break
		}
	}

	// add the notification to the list if it is not present
	if !found {
		Notifications = append(Notifications, newNotification)
	}
}

// Remove ensures that a Notification instance is removed from the list of
// controlled notifications
func Remove(cr *eventnotifierv1.SlackNotification) {

	// for each Notification in our current list
	for i, n := range Notifications {
		// check if we have this Notification already in the list
		if n.Name == cr.Name {
			// if we already have this notification in the list we want
			// to remove it
			Notifications = append(Notifications[:i], Notifications[i+1:]...)
			break
		}
	}
}

// SendMessage sends a message using all current senders
func SendEvent(event *corev1.Event, log logr.Logger) {
	// for each Notification in our current list
	for _, n := range Notifications {

		// Check if this event Kind is included in the Notification resources
		kind := false
		for _, resource := range n.Spec.Resources {
			if event.InvolvedObject.Kind == resource {
				kind = true
				break
			}
		}

		// Check if the regex matches
		re := regexp.MustCompile(n.Spec.Regex)
		regexMatched := len(re.FindAllString(event.Message, -1)) > 0

		// Send message only if both conditions are true
		if kind && regexMatched {

			// Create a Slack API client
			api := slackclient.New(n.Spec.Token)

			attachment := slackclient.Attachment{
				Fields: []slackclient.AttachmentField{
					{
						Title: "Object Kind: " + event.InvolvedObject.Kind,
					},
					{
						Title: "Object Name: " + event.InvolvedObject.Name,
					},
					{
						Title: "Namespace: " + event.InvolvedObject.Namespace,
					},
					{
						Title: "Count: " + strconv.Itoa(int(event.Count)),
					},
					{
						Title: "Reason: " + event.Reason,
					},
					{
						Title: "First Timestamp: " + event.FirstTimestamp.String(),
					},
					{
						Title: "Last Timestamp: " + event.LastTimestamp.String(),
					},
				},
			}

			// Send message to Slack
			channelID, timestamp, err := api.PostMessage(
				n.Spec.Channel,
				slackclient.MsgOptionText("*"+event.Message+"*", false),
				slack.MsgOptionAttachments(attachment),
				slackclient.MsgOptionAsUser(true),
			)

			if err != nil {
				log.Error(err, "Unable to send message to Slack")
			} else {
				log.Info("Message successfully sent to channel", "Channel", channelID, "Timestamp", timestamp)
			}
		}
	}
}
