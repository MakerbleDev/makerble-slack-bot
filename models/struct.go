package models

type JiraSearchResult struct {
	Issues []struct {
		Key    string `json:"key"`
		Fields struct {
			Summary  string                        `json:"summary"`
			Status   struct{ Name string }         `json:"status"`
			Assignee struct{ EmailAddress string } `json:"assignee"`
		} `json:"fields"`
	} `json:"issues"`
}

type Task struct {
	Key     string
	Summary string
	Status  string
}
