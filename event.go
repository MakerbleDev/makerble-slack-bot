package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"log"
	"github.com/gin-gonic/gin"
	"github.com/slack-go/slack"
)

// Handles Slack event callbacks for reactions
func handleSlackEvents(ctx *gin.Context) {
	log.Printf("Incoming Slack event request from %s", ctx.Request.RemoteAddr)
	log.Printf("Headers: %v", ctx.Request.Header)

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Printf("ERROR reading request body: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}
	log.Printf("Raw request body: %s", string(body))

	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("ERROR parsing JSON: %v | Body: %s", err, string(body))
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid JSON",
			"details": err.Error(),
		})
		return
	}
	log.Printf("Parsed payload: %+v", payload)

	if payload["type"] == "url_verification" {
		challenge, ok := payload["challenge"].(string)
		if !ok {
			log.Printf("ERROR: Challenge field missing or not a string in payload: %+v", payload)
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid challenge format",
				"details": "Challenge must be a string",
			})
			return
		}

		log.Printf("Processing URL verification challenge: %s", challenge)
		ctx.String(http.StatusOK, challenge)
		return
	}

	// Event callback handling
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
