package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

// Handles Slack event callbacks for reactions
func handleSlackEvents(ctx *gin.Context) {
	var payload map[string]interface{}
	body, _ := io.ReadAll(ctx.Request.Body)
	if err := json.Unmarshal(body, &payload); err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	if payload["type"] == "url_verification" {
		ctx.String(http.StatusOK, payload["challenge"].(string))
		return
	}

	if payload["type"] == "event_callback" {
		event := payload["event"].(map[string]interface{})
		if event["type"] == "reaction_added" && event["reaction"] == "white_check_mark" {
			userID := event["user"].(string)
			item := event["item"].(map[string]interface{})
			ts := item["ts"].(string)
			_ = sendMessageToThread(userID, ts, "Thanks for acknowledging your tasks! You’re all set for today🚀")

			user, _ := slackClient.GetUserInfo(userID)
			tasks := fetchTodaysPlanTasksFromJira()[user.Profile.Email]
			if len(tasks) > 0 {
				msg := fmt.Sprintf("*:memo: <@%s>'s Today's Plan:*\n", user.ID)

				for _, t := range tasks {
					link := fmt.Sprintf("https://makerble-backend.atlassian.net/browse/%s", t.Key)
					clickableKey := fmt.Sprintf("<%s|%s>", link, t.Key)
					msg += fmt.Sprintf("• %s — _%s_\n", clickableKey, t.Summary)
				}

				_, _, _ = slackClient.PostMessage(golangChannelID, slack.MsgOptionText(msg, false))
			} else {
				msg := fmt.Sprintf("<@%s> has no plans for today!\n", user.ID)

				_, _, _ = slackClient.PostMessage(golangChannelID, slack.MsgOptionText(msg, false))
			}
		}
	}
	ctx.Status(http.StatusOK)
}

// Sends a message to a Slack thread
func sendMessageToThread(userID, ts, msg string) error {
	_, _, err := slackClient.PostMessage(
		userID,
		slack.MsgOptionText(msg, false),
		slack.MsgOptionTS(ts),
	)
	return err
}
