package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

// This script will try and connect to the topic leader within the timeout, else return an os.Exit(1).
// We connect to the topic leader to ensure a leadership election succeeds and the topic is ready to be consumed/produced.
func main() {
	kafkaEndpoint := os.Getenv("KAFKA_ENDPOINT")
	if kafkaEndpoint == "" {
		log.Fatal("missing KAFKA_ENDPOINT environment variable")
	}

	envTopicNames := []string{"KAFKA_TOPICS_UNPROCESSED_PAYMENTS", "KAFKA_TOPICS_PAYMENT_STATE_UPDATES"}
	topics := make([]string, 0, len(envTopicNames))

	for _, tn := range envTopicNames {
		topicName := os.Getenv(tn)
		if topicName == "" {
			log.Fatal(fmt.Sprintf("missing %s environment variable", tn))
		}
		topics = append(topics, topicName)
	}

	timeout := time.After(2 * time.Minute)

	for _, topic := range topics {
		connected := false
		for {
			select {
			case <-timeout:
				log.Fatal("timeout trying to connect to kafka")
			default:
				_, err := kafka.DialLeader(context.Background(), "tcp", kafkaEndpoint, topic, 0)
				if err != nil {
					log.Println("failed to dial leader:", err)
					time.Sleep(10 * time.Second)
				} else {
					log.Printf("connected successfully to %s leader\n", topic)
					connected = true
					break
				}
			}

			if connected {
				break
			}
		}
	}

	fmt.Println("connected successfully to kafka")
}
