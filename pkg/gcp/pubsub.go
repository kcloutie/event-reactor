package gcp

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
)

func PublishEvent(projectID, topicID string, data []byte, attributes map[string]string) (string, error) {
	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return "", fmt.Errorf("failed to create the pub/sub client - %w", err)
	}
	defer client.Close()

	t := client.Topic(topicID)
	result := t.Publish(ctx, &pubsub.Message{
		Data:       data,
		Attributes: attributes,
	})
	id, err := result.Get(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to publish the message to the pub/sub topic - %w", err)
	}

	return id, nil
}
