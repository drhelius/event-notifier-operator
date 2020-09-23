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
	Log  logr.Logger
}

// Notifications is a list of current Notification items
var Notifications []Notification

// Manage ensures that a Notification instance is under control
func Manage(cr *eventnotifierv1.SlackNotification, log logr.Logger) {

	var n Notification
	n.Name = cr.Name
	n.Spec = cr.Spec.DeepCopy()
	n.Log = log

	found := false

	// for each Notification in our current list
	for _, i := range Notifications {
		// check if we have this Notification already in the list
		if n.Name == i.Name {
			// if we already have this notification in the list we want
			// to update it in case it has changed
			found = true
			i.Spec = n.Spec.DeepCopy()
			break
		}
	}

	// add the notification to the list if it is not present
	if !found {
		Notifications = append(Notifications, n)
	}
}

// SendMessage sends a message using all current senders
func SendEvent(event *corev1.Event) {
	// for each Notification in our current list
	for _, n := range Notifications {
		log := n.Log.WithName("Slack")

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

			// count: 1
			// eventTime: null
			// firstTimestamp: "2020-09-23T12:55:35Z"
			// involvedObject:
			//   apiVersion: v1
			//   fieldPath: spec.containers{order}
			//   kind: Pod
			//   name: order-v1.0.0-656c7d7457-fpb85
			//   namespace: event-notifier-operator-system
			//   resourceVersion: "58475144"
			//   uid: c58070ad-56b5-4abe-b25f-37e28d0145dd
			// kind: Event
			// lastTimestamp: "2020-09-23T12:55:35Z"
			// message: 'Readiness probe failed: dial tcp 10.129.2.35:8080: i/o timeout'
			// reason: Unhealthy
			// reportingComponent: ""
			// reportingInstance: ""
			// source:
			//   component: kubelet
			//   host: worker-2.geardome.duckdns.org

			// Send message to Slack
			channelID, timestamp, err := api.PostMessage(
				n.Spec.Channel,
				slackclient.MsgOptionText("*"+event.Message+"*", false),
				slack.MsgOptionAttachments(attachment),
				slackclient.MsgOptionAsUser(true),
			)

			log.Info("Message successfully sent to channel", "Channel", channelID, "Timestamp", timestamp)

			if err != nil {
				log.Error(err, "Unable to send message to Slack")
			}
		}
	}
}
