package models

// Summary represents a single summary document in the Summaries table.
type SummaryItem struct {
	UserID    string `dynamodbav:"user_id"`    // Partition Key
	SummaryID string `dynamodbav:"summary_id"` // Sort Key
	ServerID  string `dynamodbav:"server_id"`
	IsPrivate bool   `dynamodbav:"is_private"`
	Content   string `dynamodbav:"summary"` // Using "summary" tag to match original field name
	CreatedAt string `dynamodbav:"created_at"`
	UpdatedAt string `dynamodbav:"updated_at"`
}
