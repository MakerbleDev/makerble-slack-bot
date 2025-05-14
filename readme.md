This is a Slackbot designed to send acknowledgment messages to all users in the makerble-backend Jira organization. It helps track task acknowledgment based on users' daily plan status in Jira.  

---

1. Daily Acknowledgment Messages
   - Sends an acknowledgment message every day at 9 AM to all users.  
   - Users have until the next day (9 AM) to acknowledge their tasks.  

2. Task Status Integration
   - Fetches task statuses from Jira and checks for user acknowledgments.  

3. Reaction Capture
   - Captures Slack message reactions via event subscriptions.  

## Important
   - You must expose a public HTTPS endpoint (e.g., `https://your-domain.com/slack/events`)
      This endpoint must be added to:  
      Slack App Settings → Event Subscriptions → Request URL

- Remember to renew Jira API token every year

- Users must react to messages for the bot to track acknowledgements

## Run it
```bash
go run .