package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	// ADDED: AWS SDK v2 imports for configuration and DynamoDB service.
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	// REMOVED: MongoDB driver imports are no longer needed.
	// "go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectDB initializes and returns a new DynamoDB client.
// The AWS SDK v2 will automatically look for credentials in the environment
// (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY), shared credentials file (~/.aws/credentials),
// or an IAM role if running on AWS infrastructure.
func ConnectDB() *dynamodb.Client {
	// CHANGED: We now look for AWS_REGION instead of MONGO_URI.
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		log.Fatal("AWS_REGION not set in environment variable")
	}
	fmt.Println("AWS Region:", awsRegion)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	fmt.Println("Loading AWS configuration...")
	// The LoadDefaultConfig function is the standard way to load configuration.
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		log.Fatalf("Unable to load AWS SDK config: %v", err)
	}

	// Create a new DynamoDB client from the loaded configuration.
	client := dynamodb.NewFromConfig(cfg)

	// Optional but recommended: Perform a simple, low-cost operation to verify
	// that the credentials and region are correct. This is the equivalent of a "ping".
	fmt.Println("Verifying connection to DynamoDB...")
	_, err = client.ListTables(ctx, &dynamodb.ListTablesInput{})
	if err != nil {
		log.Fatalf("Failed to connect to DynamoDB. Check credentials and region. Error: %v", err)
	}

	fmt.Println("Connected to DynamoDB successfully")
	return client
}

// REMOVED: The DisconnectDB function is not needed for the AWS SDK v2.
// The SDK manages underlying HTTP connections in its connection pool and does not
// maintain a persistent stateful connection that needs to be explicitly closed.
