## Kafka Adapter 

[TOC]

Kafka adapter its a interface in front of Shopify/Sarama go package, producer part is easy to use but for consumer you need to copy the example code and change it to your need.

#### Env variables

```
NOTIFICATION_EMAIL_SMTP_HOST="faza.io"
NOTIFICATION_EMAIL_SMTP_PORT="587"
NOTIFICATION_EMAIL_SMTP_USER="SMTP_USER"
NOTIFICATION_EMAIL_SMTP_PASS="SMTP_PASS"

// APP MODE: consumer, server
NOTIFICATION_APP_MODE="consumer"
NOTIFICATION_APP_PORT="8080"

NOTIFICATION_KAFKA_VERSION="2.2.0"
NOTIFICATION_KAFKA_BROKERS="localhost:9092,localhost:9093,localhost:9094"
// Consumer Group Name
NOTIFICATION_KAFKA_CONSUMER_GROUP="notification-service"
// Topics: 
// notification-email-batch 
// notification-email-single
// notification-sms-batch
// notification-sms-single
NOTIFICATION_KAFKA_CONSUMER_TOPIC="notification-email-batch"
```





producer support two methods `sendOne` and `SendBulk`

#### producer example (Send On):

```go
package main

import (
	"fmt"

	"gitlab.faza.io/go-framework/kafkaadapter"
)

func main() {
	k := kafkaadapter.NewKafka([]string{"localhost:9092", "localhost:9093"}, "test-topic")
	k.Config.Producer.Return.Successes = true
	
	p, o, err := k.SendOne("", []byte("Hello world!"))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Partition:", p, " Offset:", o)
}
```



#### producer example (Send Bulk):

```go
package main

import (
	"fmt"

	"gitlab.faza.io/go-framework/kafkaadapter"
)

func main() {
	k := kafkaadapter.NewKafka([]string{"localhost:9092", "localhost:9093"}, "test-topic")
	k.Config.Producer.Return.Successes = true
	
	for i := 0; i <= 20; i++ {
		k.AddToBulk("", []byte("Hello world! "))
	}

	err := k.SendBulk()
	if err != nil {
		fmt.Println(err)
	}
}
```



#### producer example (Send json):

```go
package main

import (
	"fmt"

	"gitlab.faza.io/go-framework/kafkaadapter"
)

func main() {
	k := NewKafka([]string{"localhost:9092", "localhost:9093"}, "test-topic")
	k.Config.Producer.Return.Successes = true

	type TestMessage struct {
		Message string
	}
	e := TestMessage{
		Message: "test message to send to kafka",
	}

	message, err := json.Marshal(e)
	if err != nil {
		log.Fatal(err)
	}
	
	p, o, err := k.SendOne("", message)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Partition:", p, " Offset:", o)
}
```





#### consumer example:

you can write your business login in  ```ConsumeClaim```  function 

```session.MarkMessage(message, "")``` is the function that handle marking the topic offset as consumed  

```go
import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/Shopify/sarama"
)

func (k *Kafka) ConsumerGroup(kafkaVersion string) error {
	log.Println("Starting a new kafka consumer")

	sarama.Logger = log.New(os.Stdout, "[FAZA.io] ", log.LstdFlags)

	version, err := sarama.ParseKafkaVersion(kafkaVersion)
	if err != nil {
		return errors.New("Error parsing Kafka version: " + err.Error())
	}
	config := sarama.NewConfig()
	config.Version = version

	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer := Consumer{
		ready: make(chan bool, 0),
	}

	ctx, cancel := context.WithCancel(context.Background())
	client, err := sarama.NewConsumerGroup(k.brokers, k.topic, config)
	if err != nil {
		return errors.New("Error creating consumer group client: " + err.Error())
	}

	wg := &sync.WaitGroup{}
	go func() {
		wg.Add(1)
		defer wg.Done()
		for {
			if err := client.Consume(ctx, []string{k.topic}, &consumer); err != nil {
				return
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool, 0)
		}
	}()

	<-consumer.ready // Await till the consumer has been set up
	log.Println("Sarama consumer up and running!...")

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
		return errors.New("Error closing client: " + err.Error())
	}
	return nil
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	ready chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		// @TODO: Write your business logic here
		log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		session.MarkMessage(message, "")
	}

	return nil
}

```


