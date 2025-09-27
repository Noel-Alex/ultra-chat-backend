package models

// User defines the structure for a user in the application.
// The `dynamodbav` struct tags are used by the AWS SDK to marshal/unmarshal
// Go structs to and from DynamoDB items.
type User struct {
	// ID will be the Partition Key for our DynamoDB table. It's the unique
	// identifier for each user, coming from Discord's user ID.
	ID            string                 `dynamodbav:"id"`
	UUID          string                 `dynamodbav:"uuid"`
	Token         map[string]interface{} `dynamodbav:"token"` // This will be stored as a DynamoDB Map type.
	Username      string                 `dynamodbav:"username"`
	Discriminator string                 `dynamodbav:"discriminator"`

	// Summaries will be stored as a DynamoDB List of Maps. The `omitempty` tag
	// means this field will not be written to DynamoDB if the slice is empty or nil.
	Summaries []Summary `dynamodbav:"summaries,omitempty"`
}

// Summary defines a nested object within the User item.
// It does not need its own primary key as it's part of the User item.
type Summary struct {
	ID      string `dynamodbav:"id"`
	Content string `dynamodbav:"content"`
}
