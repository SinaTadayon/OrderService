package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"gitlab.faza.io/go-framework/logger"

	"github.com/Shopify/sarama"
)

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *PayToSellerFailedConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for m := range claim.Messages() {
		logger.Audit("PayToSellerFailed: value = %s, OffSet = %v, topic = %s",
			string(m.Value), m.Offset, m.Topic)
		// Validate message
		message, err := PayToSellerFailedMessageValidate(m)
		if err != nil {
			logger.Err("PayToSellerFailed validation failed: %s", err)
			continue
		}
		// Message action
		err = PayToSellerFailedAction(message)
		if err != nil {
			logger.Err("PayToSellerFailed action failed: %s", err)
			continue
		}
		session.MarkMessage(message, "")
	}
	return nil
}

func startPayToSellerFailed(Version string, topics string) {
	logger.Audit("starting PayToSellerFailed consumers...")
	if App.config.Kafka.Brokers == "" {
		log.Fatal("Cant start PayToSellerFailed consumer, No brokers defined")
	}
	if App.config.Kafka.ConsumerTopic == "" {
		log.Fatal("Cant start PayToSellerFailed consumer, No topic defined")
	}
	if App.config.Kafka.Version == "" {
		log.Fatal("Cant start PayToSellerFailed consumer, No kafka version defined")
	}

	brokers := strings.Split(App.config.Kafka.Brokers, ",")
	version, err := sarama.ParseKafkaVersion(Version)
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}

	config := sarama.NewConfig()
	config.Version = version
	//config.Consumer.Retry.Backoff
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer := PayToSellerFailedConsumer{
		ready: make(chan bool, 0),
	}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(brokers, App.config.Kafka.ConsumerGroup, config)
	if err != nil {
		log.Panicf("Error creating PayToSellerFailed consumer group client: %v", err)
	}

	wg := &sync.WaitGroup{}
	go func() {
		wg.Add(1)
		defer wg.Done()
		for {
			if err := client.Consume(ctx, strings.Split(topics, ","), &consumer); err != nil {
				log.Panicf("Error from PayToSellerFailed consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool, 0)
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	log.Println("PayToSellerFailed consumer up and running!...")

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		log.Println("terminating: context cancelled")
	case <-sigterm:
		log.Println("terminating: via signal")
	}
	cancel()
	wg.Wait()
	if err = client.Close(); err != nil {
		log.Panicf("Error closing client: %v", err)
	}
}

// Consumer represents a Sarama consumer group consumer
type PayToSellerFailedConsumer struct {
	ready chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *PayToSellerFailedConsumer) Setup(sarama.ConsumerGroupSession) error {
	close(c.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *PayToSellerFailedConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}
