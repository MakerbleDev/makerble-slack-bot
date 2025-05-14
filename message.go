package main

import (
	"fmt"
	"slack-bot/logger"
	"slack-bot/models"

	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

// Sends a task message to a user
func sendTaskMessage(userID string, tasks []models.Task) {
	var msg string
	if len(tasks) > 0 {
		msg = fmt.Sprintf(
			"Hello <@%s>,\n"+
				"*Acknowledge Your Today's Plan:*\n",
			userID,
		)
		for _, t := range tasks {
			link := fmt.Sprintf("https://makerble-backend.atlassian.net/browse/%s", t.Key)
			clickableKey := fmt.Sprintf("<%s|%s>", link, t.Key)
			msg += fmt.Sprintf("• %s — %s\n", clickableKey, t.Summary)
		}
		msg += "\n" +
			"*Reminder:* Before acknowledging, ensure that all your tasks for today's plan are added. " +
			"React with :white_check_mark: to *acknowledge*."
	}
	postMessageToSlack(userID, msg, false)
}

// Sends a no-task message to a user
func sendNoTaskMessage(userID string) {
	msg := fmt.Sprintf(
		"Hello <@%s>,\n"+
			"*It looks like you don't have any tasks under 'Today's Plan' in Jira.*\n"+
			"To stay on track, please update your tasks for today in the Jira board.\n"+
			"Once updated, React with :white_check_mark: to *acknowledge*.\n",
		userID,
	)
	postMessageToSlack(userID, msg, false)
}

func postMessageToSlack(userID, msg string, false bool) {
	_, _, err := slackClient.PostMessage(userID, slack.MsgOptionText(msg, false))
	if err != nil {
		logger.Log.Warn("Message send failed", zap.Error(err), zap.String("user", userID))
	}
}