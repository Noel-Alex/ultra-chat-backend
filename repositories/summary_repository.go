package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"ultra-chat-backend/models"
)

// SummaryRepository defines the interface for summary data operations.
// Note the more specific function signatures compared to the generic bson.M.
type SummaryRepository interface {
	AddSummary(summary models.SummaryItem) error
	GetSummariesByUserID(userID string) ([]models.SummaryItem, error)
	GetSummariesByUserIDAndServerID(userID, serverID string) ([]models.SummaryItem, error)
	UpdateSummaryContent(userID, summaryID, content string) error
	DeleteSummary(userID, summaryID string) error
	CheckUserExists(userID string) (bool, error)
}

// dynamoDBSummaryRepository is the DynamoDB implementation of SummaryRepository.
type dynamoDBSummaryRepository struct {
	ddbClient          *dynamodb.Client
	summariesTableName string
	usersTableName     string
}

// NewSummaryRepository initializes the repository with a DynamoDB client.
// Infrastructure tasks like creating tables and indexes should be done
// outside of the application code (e.g., using CloudFormation, CDK, or Terraform).
func NewSummaryRepository(client *dynamodb.Client) SummaryRepository {
	return &dynamoDBSummaryRepository{
		ddbClient:          client,
		summariesTableName: "Summaries", // Manage via config
		usersTableName:     "Users",     // Manage via config
	}
}

// AddSummary uses PutItem to insert a new summary.
func (r *dynamoDBSummaryRepository) AddSummary(summary models.SummaryItem) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	item, err := attributevalue.MarshalMap(summary)
	if err != nil {
		return fmt.Errorf("failed to marshal summary: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.summariesTableName),
		Item:      item,
	}

	_, err = r.ddbClient.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to add summary to dynamodb: %w", err)
	}
	return nil
}

// GetSummariesByUserID queries the base table using the user_id partition key.
func (r *dynamoDBSummaryRepository) GetSummariesByUserID(userID string) ([]models.SummaryItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	keyEx := expression.Key("user_id").Equal(expression.Value(userID))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build query expression: %w", err)
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(r.summariesTableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	}

	result, err := r.ddbClient.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to query summaries by user id: %w", err)
	}

	var summaries []models.SummaryItem
	err = attributevalue.UnmarshalListOfMaps(result.Items, &summaries)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal summaries: %w", err)
	}
	return summaries, nil
}

// GetSummariesByUserIDAndServerID queries the GSI using user_id and server_id.
func (r *dynamoDBSummaryRepository) GetSummariesByUserIDAndServerID(userID, serverID string) ([]models.SummaryItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	keyEx := expression.Key("user_id").Equal(expression.Value(userID)).
		And(expression.Key("server_id").Equal(expression.Value(serverID)))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build GSI query expression: %w", err)
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(r.summariesTableName),
		IndexName:                 aws.String("UserServerIndex"), // Querying the GSI
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	}

	result, err := r.ddbClient.Query(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to GSI query summaries: %w", err)
	}

	var summaries []models.SummaryItem
	err = attributevalue.UnmarshalListOfMaps(result.Items, &summaries)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal GSI summaries: %w", err)
	}
	return summaries, nil
}

// UpdateSummaryContent uses UpdateItem to modify a specific summary.
// Note: This requires both user_id and summary_id to uniquely identify the item.
// The handler will need to fetch the summary first to get its summary_id if it doesn't have it.
func (r *dynamoDBSummaryRepository) UpdateSummaryContent(userID, summaryID, content string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := map[string]types.AttributeValue{
		"user_id":    &types.AttributeValueMemberS{Value: userID},
		"summary_id": &types.AttributeValueMemberS{Value: summaryID},
	}

	update := expression.Set(expression.Name("summary"), expression.Value(content)).
		Set(expression.Name("updated_at"), expression.Value(time.Now().UTC().Format(time.RFC3339)))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return fmt.Errorf("failed to build update expression: %w", err)
	}

	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.summariesTableName),
		Key:                       key,
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ConditionExpression:       aws.String("attribute_exists(user_id)"), // Ensures we don't create an item
	}

	_, err = r.ddbClient.UpdateItem(ctx, input)
	if err != nil {
		if _, ok := err.(*types.ConditionalCheckFailedException); ok {
			return errors.New("no matching summary found")
		}
		return fmt.Errorf("failed to update summary: %w", err)
	}
	return nil
}

// DeleteSummary uses DeleteItem with the full primary key.
func (r *dynamoDBSummaryRepository) DeleteSummary(userID, summaryID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := map[string]types.AttributeValue{
		"user_id":    &types.AttributeValueMemberS{Value: userID},
		"summary_id": &types.AttributeValueMemberS{Value: summaryID},
	}

	input := &dynamodb.DeleteItemInput{
		TableName:           aws.String(r.summariesTableName),
		Key:                 key,
		ConditionExpression: aws.String("attribute_exists(user_id)"), // Ensures it exists before deleting
	}

	_, err := r.ddbClient.DeleteItem(ctx, input)
	if err != nil {
		if _, ok := err.(*types.ConditionalCheckFailedException); ok {
			return errors.New("no matching summary found")
		}
		return fmt.Errorf("failed to delete summary: %w", err)
	}
	return nil
}

// CheckUserExists uses GetItem on the Users table.
func (r *dynamoDBSummaryRepository) CheckUserExists(userID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.usersTableName),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: userID},
		},
	}

	result, err := r.ddbClient.GetItem(ctx, input)
	if err != nil {
		return false, err
	}
	// If result.Item is not nil, the user exists.
	return result.Item != nil, nil
}
