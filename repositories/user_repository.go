package repositories

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"log"
	"time"
	"ultra-chat-backend/models"
)

// UserRepository interface is now database-agnostic.
// We've replaced bson.M with generic Go types like map[string]interface{} and models.Summary.
type UserRepository interface {
	FindUserByID(id string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(id string, updateData map[string]interface{}) error
	AddSummary(userID string, summary models.Summary) error
	GetSummaries(userID string) ([]models.Summary, error)
	UpdateSummary(userID, summaryID, content string) error
	DeleteSummary(userID, summaryID string) error
	IsAuthenticated(userID string) (bool, error)
}

// userRepository now holds a DynamoDB client and the table name.
type userRepository struct {
	ddbClient *dynamodb.Client
	tableName string
}

// NewUserRepository is the constructor for our DynamoDB repository.
// It takes a configured DynamoDB client as input.
func NewUserRepository(client *dynamodb.Client) UserRepository {
	return &userRepository{
		ddbClient: client,
		tableName: "Users", // It's good practice to manage this via config/env variables.
	}
}

// FindUserByID uses GetItem to retrieve a user by their partition key.
func (r *userRepository) FindUserByID(id string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	key, err := attributevalue.Marshal(id)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user ID: %w", err)
	}

	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableName),
		Key: map[string]types.AttributeValue{
			"id": key,
		},
	}

	result, err := r.ddbClient.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get item from DynamoDB: %w", err)
	}

	// If no item is found, result.Item will be nil. This is the DynamoDB
	// equivalent of mongo.ErrNoDocuments.
	if result.Item == nil {
		return nil, errors.New("user not found")
	}

	var user models.User
	if err = attributevalue.UnmarshalMap(result.Item, &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	return &user, nil
}

// CreateUser uses PutItem to create a new user item.
func (r *userRepository) CreateUser(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user for creation: %w", err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableName),
		Item:      item,
	}

	_, err = r.ddbClient.PutItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put item to DynamoDB: %w", err)
	}
	return nil
}

// UpdateUser uses UpdateItem to modify specific fields of a user.
func (r *userRepository) UpdateUser(id string, updateData map[string]interface{}) error {
	// <<< --- ADD THIS CHECK --- >>>
	// If the map of data to update is empty, there's nothing to do.
	// This prevents the expression builder from failing.
	if len(updateData) == 0 {
		log.Printf("UpdateUser called for user %s with no data to update. Skipping.", id)
		return nil // Return success, as no work was needed.
	}
	// <<< --- END OF ADDED CHECK --- >>>

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Dynamically build the update expression based on the map provided.
	var updateBuilder expression.UpdateBuilder
	for k, v := range updateData {
		updateBuilder.Set(expression.Name(k), expression.Value(v))
	}

	expr, err := expression.NewBuilder().WithUpdate(updateBuilder).Build()
	if err != nil {
		return fmt.Errorf("failed to build update expression: %w", err)
	}

	key, err := attributevalue.Marshal(id)
	if err != nil {
		return fmt.Errorf("failed to marshal user ID for update: %w", err)
	}

	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.tableName),
		Key:                       map[string]types.AttributeValue{"id": key},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		ReturnValues:              types.ReturnValueNone,
	}

	_, err = r.ddbClient.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to update user in DynamoDB: %w", err)
	}
	return nil
}

// AddSummary uses UpdateItem with list_append to add a new summary to the list.
func (r *userRepository) AddSummary(userID string, summary models.Summary) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	key, err := attributevalue.Marshal(userID)
	if err != nil {
		return fmt.Errorf("failed to marshal user ID for adding summary: %w", err)
	}

	// We need to wrap the single summary in a slice to append it to the list.
	summaryList := []models.Summary{summary}

	// THIS IS THE CORRECTED SECTION
	// 1. We use expression.ListAppend, not expression.Call.
	// 2. The first argument to ListAppend handles the case where the 'summaries' attribute
	//    doesn't exist yet, by creating it with an empty list. This perfectly
	//    mimics MongoDB's `$push` behavior.
	// 3. The second argument is the new list of items we want to append.
	update := expression.Set(
		expression.Name("summaries"),
		expression.ListAppend(
			expression.IfNotExists(expression.Name("summaries"), expression.Value([]interface{}{})),
			expression.Value(summaryList),
		),
	)

	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		return fmt.Errorf("failed to build expression for adding summary: %w", err)
	}

	input := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.tableName),
		Key:                       map[string]types.AttributeValue{"id": key},
		UpdateExpression:          expr.Update(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	_, err = r.ddbClient.UpdateItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to add summary in DynamoDB: %w", err)
	}
	return nil
}

// GetSummaries uses a Projection Expression to fetch only the summaries list.
func (r *userRepository) GetSummaries(userID string) ([]models.Summary, error) {
	user, err := r.FindUserByID(userID)
	if err != nil {
		return nil, err
	}
	// The User model already contains the summaries, so we just return them.
	// This is simpler than using a projection expression and keeps FindUserByID reusable.
	return user.Summaries, nil
}

// To update/delete a specific item in a list, DynamoDB requires its index.
// The common pattern is: Read the list, find the index, then send an update command.
func (r *userRepository) findSummaryIndex(userID, summaryID string) (int, *models.User, error) {
	user, err := r.FindUserByID(userID)
	if err != nil {
		return -1, nil, err
	}

	for i, summary := range user.Summaries {
		if summary.ID == summaryID {
			return i, user, nil
		}
	}

	return -1, user, errors.New("summary not found")
}

// UpdateSummary finds the summary's index and updates its content field.
func (r *userRepository) UpdateSummary(userID, summaryID, content string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	index, _, err := r.findSummaryIndex(userID, summaryID)
	if err != nil {
		return err
	}

	key, _ := attributevalue.Marshal(userID)
	newContent, _ := attributevalue.Marshal(content)

	// Here we use an UpdateExpression that targets a specific element in the list by its index.
	// SET summaries[<index>].content = :new_content
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key:       map[string]types.AttributeValue{"id": key},
		UpdateExpression: aws.String(
			fmt.Sprintf("SET #summaries[%d].#content = :content", index),
		),
		ExpressionAttributeNames: map[string]string{
			"#summaries": "summaries",
			"#content":   "content",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":content": newContent,
		},
	}

	_, err = r.ddbClient.UpdateItem(ctx, input)
	return err
}

// DeleteSummary finds the summary's index and uses the REMOVE action.
func (r *userRepository) DeleteSummary(userID, summaryID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	index, _, err := r.findSummaryIndex(userID, summaryID)
	if err != nil {
		return err
	}

	key, _ := attributevalue.Marshal(userID)

	// The REMOVE action can delete a specific element from a list by its index.
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(r.tableName),
		Key:       map[string]types.AttributeValue{"id": key},
		UpdateExpression: aws.String(
			fmt.Sprintf("REMOVE #summaries[%d]", index),
		),
		ExpressionAttributeNames: map[string]string{
			"#summaries": "summaries",
		},
	}

	_, err = r.ddbClient.UpdateItem(ctx, input)
	return err
}

// IsAuthenticated checks for the existence of a user.
func (r *userRepository) IsAuthenticated(userID string) (bool, error) {
	_, err := r.FindUserByID(userID)
	if err != nil {
		// If the error is "user not found", it means they don't exist. Not a real error.
		if err.Error() == "user not found" {
			return false, nil
		}
		// Any other error is a system error.
		return false, err
	}
	// If no error, the user was found.
	return true, nil
}
