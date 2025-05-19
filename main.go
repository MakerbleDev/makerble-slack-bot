package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slack-bot/logger"
	"slack-bot/models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
	"go.uber.org/zap"
)

var slackClient *slack.Client
var golangChannelID, leaveChannelID, jiraEmail, jiraToken, jiraBaseURL string

func main() {
	_ = godotenv.Load()
	logger.InitLogger()

	slackClient = slack.New(os.Getenv("SLACK_TOKEN"))
	golangChannelID = os.Getenv("SLACK_GOLANG_CHANNEL_ID")
	leaveChannelID = os.Getenv("SLACK_LEAVE_CHANNEL_ID")
	jiraEmail = os.Getenv("JIRA_EMAIL")
	jiraToken = os.Getenv("JIRA_API_TOKEN")
	jiraBaseURL = os.Getenv("JIRA_BASE_URL")

	authTest()

	loc, _ := time.LoadLocation("Asia/Kolkata")
	c := cron.New(cron.WithLocation(loc))
	_, err := c.AddFunc("0 9 * * *", sendAcknowledgmentMessagesForAllUsers)
	if err != nil {
		logger.Log.Fatal("Cron setup error", zap.Error(err))
	}
	c.Start()

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.POST("/events", handleSlackEvents)
	logger.Log.Info("Server running on port 3000")
	_ = r.Run("0.0.0.0:3000")
}

// Tests Slack bot authorization
func authTest() {
	authTest, err := slackClient.AuthTest()
	if err != nil {
		logger.Log.Fatal("Slack token is invalid or Slack API error", zap.Error(err))
	}
	logger.Log.Info(fmt.Sprintf("Slack bot connected as: %s (Team: %s)", authTest.User, authTest.Team))
}

// Fetches Slack users who belong to the Jira organization
func fetchJiraOrgSlackUsers() []*slack.User {
	users, err := slackClient.GetUsers()
	if err != nil {
		logger.Log.Warn("Failed to fetch Slack users", zap.Error(err))
		return nil
	}

	// Fetch Jira organization emails
	jiraOrgEmails := fetchJiraOrgEmails()
	if jiraOrgEmails == nil {
		logger.Log.Warn("Failed to fetch Jira organization emails")
		return nil
	}

	// Convert []slack.User to []*slack.User, filtering by Jira organization emails
	userPtrs := []*slack.User{}
	jiraEmailSet := make(map[string]struct{}, len(jiraOrgEmails))
	for _, email := range jiraOrgEmails {
		jiraEmailSet[email] = struct{}{}
	}

	for i := range users {
		if _, exists := jiraEmailSet[users[i].Profile.Email]; exists {
			userPtrs = append(userPtrs, &users[i])
		}
	}

	return userPtrs
}

// Fetches Jira organization emails
func fetchJiraOrgEmails() []string {
	url := fmt.Sprintf("%s/rest/api/3/users/search", jiraBaseURL)

	req, _ := http.NewRequest("GET", url, nil)
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", jiraEmail, jiraToken)))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		logger.Log.Warn("Failed to fetch Jira users", zap.Error(err))
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var jiraUsers []struct {
		EmailAddress string `json:"emailAddress"`
	}
	if err := json.Unmarshal(body, &jiraUsers); err != nil {
		logger.Log.Warn("JSON parse error for Jira users", zap.Error(err))
		return nil
	}

	// Extract email addresses
	emails := make([]string, 0, len(jiraUsers))
	for _, user := range jiraUsers {
		emails = append(emails, user.EmailAddress)
	}
	return emails
}

// Fetches Jira tasks for "Today's Plan" status from Jira API
func fetchTodaysPlanTasksFromJira() map[string][]models.Task {
	jql := `status="Today's Plan" AND assignee IS NOT EMPTY`
	url := fmt.Sprintf("%s/rest/api/2/search?jql=%s", jiraBaseURL, strings.ReplaceAll(jql, " ", "%20"))

	req, _ := http.NewRequest("GET", url, nil)
	auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", jiraEmail, jiraToken)))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		logger.Log.Warn("Failed Jira fetch", zap.Error(err))
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result models.JiraSearchResult
	if err := json.Unmarshal(body, &result); err != nil {
		logger.Log.Warn("JSON parse error", zap.Error(err))
		return nil
	}

	tasksByUser := make(map[string][]models.Task)
	for _, issue := range result.Issues {
		email := issue.Fields.Assignee.EmailAddress
		task := models.Task{
			Key:     issue.Key,
			Summary: issue.Fields.Summary,
			Status:  issue.Fields.Status.Name,
		}
		tasksByUser[email] = append(tasksByUser[email], task)
	}
	return tasksByUser
}
