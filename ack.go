package main

import (
	"fmt"
	"slack-bot/logger"
	"strings"
	"time"

	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

func checkAcknowledgmentsFromPreviousRun(usersToCheck map[string]struct {
	slackID  string
	hasTasks bool
}) {
	logger.Log.Info("Starting acknowledgment check")
	now := time.Now()
	windowStart := now.Add(-24 * time.Hour) // 9 AM previous day
	windowEnd := now

	logger.Log.Info("Time window for checking history",
		zap.String("start", windowStart.Format("15:04:05")),
		zap.String("end", windowEnd.Format("15:04:05")),
	)

	for email, _ := range usersToCheck {
		logger.Log.Info("Processing user", zap.String("email", email))

		// Get user info
		user, err := slackClient.GetUserByEmail(email)
		if err != nil {
			logger.Log.Warn("Failed to fetch user by email", zap.String("email", email), zap.Error(err))
			continue
		}
		logger.Log.Info("User found", zap.String("userID", user.ID))

		// Open conversation
		channel, _, _, err := slackClient.OpenConversation(&slack.OpenConversationParameters{
			Users: []string{user.ID},
		})
		if err != nil {
			logger.Log.Warn("Failed to open conversation", zap.String("userID", user.ID), zap.Error(err))
			continue
		}
		logger.Log.Info("Conversation opened", zap.String("channelID", channel.ID))

		// Get conversation history
		history, err := slackClient.GetConversationHistory(&slack.GetConversationHistoryParameters{
			ChannelID: channel.ID,
			Oldest:    fmt.Sprintf("%d", windowStart.Unix()),
			Latest:    fmt.Sprintf("%d", windowEnd.Unix()),
			Limit:     1,
		})
		if err != nil {
			logger.Log.Warn("Failed to fetch conversation history", zap.String("channelID", channel.ID), zap.Error(err))
			continue
		}

		if len(history.Messages) == 0 {
			logger.Log.Info("No messages found in the given time window", zap.String("channelID", channel.ID))
		} else {
			logger.Log.Info("Messages found", zap.Int("count", len(history.Messages)))
		}

		// Check for acknowledgment
		acknowledged := false
		for _, msg := range history.Messages {
			if strings.Contains(msg.Text, "React with :white_check_mark:") {
				logger.Log.Info("Found acknowledgment request message")

				itemRef := slack.ItemRef{
					Channel:   channel.ID,
					Timestamp: msg.Timestamp,
				}
				reactions, err := slackClient.GetReactions(itemRef, slack.GetReactionsParameters{Full: true})
				if err != nil {
					logger.Log.Warn("Failed to fetch reactions", zap.Error(err))
					continue
				}

				for _, reaction := range reactions {
					if reaction.Name == "white_check_mark" {
						for _, userID := range reaction.Users {
							if userID == user.ID {
								logger.Log.Info("Acknowledgment received", zap.String("userID", user.ID))
								acknowledged = true
								break
							}
						}
					}
					if acknowledged {
						break
					}
				}
			}
		}

		if !acknowledged {
			msg := fmt.Sprintf(
				":warning: *Pending Acknowledgment*\n"+
					"<@%s> has not acknowledged the required message from yesterday. Please follow up with them if needed to ensure it's addressed.\n"+
					"Thank you! :pray:",
				user.ID,
			)
			logger.Log.Info("Posting to leave channel", zap.String("channelID", leaveChannelID))
			postMessageToSlack(leaveChannelID, msg, false)
		} else {
			logger.Log.Info("No action needed; acknowledgment status verified")
		}

	}

	logger.Log.Info("Acknowledgment check completed")
}

func sendAcknowledgmentMessagesForAllUsers() {
	tasksByUser := fetchTodaysPlanTasksFromJira()
	allUsers := fetchJiraOrgSlackUsers()

	usersToCheck := make(map[string]struct {
		slackID  string
		hasTasks bool
	})

	for email, tasks := range tasksByUser {
		user, err := slackClient.GetUserByEmail(email)
		if err != nil {
			logger.Log.Debug("Slack user not found", zap.String("email", email))
			continue
		}
		usersToCheck[email] = struct {
			slackID  string
			hasTasks bool
		}{user.ID, true}
		sendTaskMessage(user.ID, tasks)
	}

	// Add users without tasks
	for _, user := range allUsers {
		email := user.Profile.Email
		if email == "" {
			logger.Log.Warn("User email missing", zap.String("userID", user.ID))
			continue
		}
		if _, exists := tasksByUser[email]; !exists {
			usersToCheck[email] = struct {
				slackID  string
				hasTasks bool
			}{user.ID, false}
			sendNoTaskMessage(user.ID)
		}
	}

	logger.Log.Info("Total Users to Check Summary",
		zap.Int("Total", len(usersToCheck)),
		zap.Int("With Tasks", len(tasksByUser)),
		zap.Int("Without Tasks", len(usersToCheck)-len(tasksByUser)),
	)

	checkAcknowledgmentsFromPreviousRun(usersToCheck)
}
