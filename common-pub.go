package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
)

var (
	topic *pubsub.Topic

	// Messages received by this instance.
	messagesMu sync.Mutex
	messages   []string
)

type pushRequest struct {
	Message struct {
		Attributes map[string]string
		Data       []byte
		ID         string `json:"message_id"`
	}
	Subscription string
}

type NetChangeTrigger struct {
	OrgId      string
	ItemId     string
	LocationId string
}

func publishTrigger(triggerType string, triggers []NetChangeTrigger) error {

	var err error
	var topicName string

	if triggerType == "forecast" {
		topicName = mustGetenv("PUBSUB_TOPIC_FORECAST")
	} else if triggerType == "supply" {
		topicName = mustGetenv("PUBSUB_TOPIC_SUPPLY")
	} else if triggerType == "route" {
		topicName = mustGetenv("PUBSUB_TOPIC_ROUTE")
	}

	ctx := context.Background()
	client, err := pubsub.NewClient(ctx, mustGetenv("GOOGLE_CLOUD_PROJECT"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	topic = client.Topic(topicName)

	j, err := json.Marshal(triggers)
	if err != nil {
		fmt.Println("Error converting trigger slice to JSON.\n[ERROR] -", err)
		return err
	}

	msg := &pubsub.Message{
		Data: []byte(j),
		ID:   "",
	}

	var result *pubsub.PublishResult
	result = topic.Publish(ctx, msg)
	ID, err := result.Get(ctx)

	if err != nil {
		//TODO remove testing
		fmt.Println("error topic.Publish:", err)
		return err
	}
	//TODO remove testing
	fmt.Println("published message ", topicName, "ID ", ID)
	fmt.Println("triggers:", triggers)

	return err
}

func contains(triggers []NetChangeTrigger, new NetChangeTrigger) bool {
	for _, existing := range triggers {
		if existing.OrgId == new.OrgId &&
			existing.ItemId == new.ItemId &&
			existing.LocationId == new.LocationId {
			return true
		}
	}
	return false
}
