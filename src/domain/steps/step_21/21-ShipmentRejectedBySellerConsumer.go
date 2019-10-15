package shipment_rejected_by_seller_step


//
//import (
//	"context"
//	"gitlab.faza.io/order-project/order-service"
//	"log"
//	"os"
//	"os/signal"
//	"strings"
//	"sync"
//	"syscall"
//
//	"gitlab.faza.io/go-framework/logger"
//
//	"github.com/Shopify/sarama"
//)
//
//// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
//func (c *ShipmentRejectedBySellerConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
//	for m := range claim.Messages() {
//		logger.Audit("ShipmentRejectedBySeller: value = %s, OffSet = %v, topic = %s",
//			string(m.Value), m.Offset, m.Topic)
//		// Validate message
//		message, err := ShipmentRejectedBySellerMessageValidate(m)
//		if err != nil {
//			logger.Err("ShipmentRejectedBySeller validation failed: %s", err)
//			continue
//		}
//		// Message action
//		err = ShipmentRejectedBySellerAction(message)
//		if err != nil {
//			logger.Err("ShipmentRejectedBySeller action failed: %s", err)
//			continue
//		}
//		session.MarkMessage(message, "")
//	}
//	return nil
//}
//
//func startShipmentRejectedBySeller(Version string, topics string) {
//	logger.Audit("starting ShipmentRejectedBySeller consumers...")
//	if main.App.Config.Kafka.Brokers == "" {
//		log.Fatal("Cant start ShipmentRejectedBySeller consumer, No brokers defined")
//	}
//	if main.App.Config.Kafka.ConsumerTopic == "" {
//		log.Fatal("Cant start ShipmentRejectedBySeller consumer, No topic defined")
//	}
//	if main.App.Config.Kafka.Version == "" {
//		log.Fatal("Cant start ShipmentRejectedBySeller consumer, No kafka version defined")
//	}
//
//	brokers := strings.Split(main.App.Config.Kafka.Brokers, ",")
//	version, err := sarama.ParseKafkaVersion(Version)
//	if err != nil {
//		log.Panicf("Error parsing Kafka version: %v", err)
//	}
//
//	config := sarama.NewConfig()
//	config.Version = version
//	//Config.Consumer.Retry.Backoff
//	config.Consumer.Offsets.Initial = sarama.OffsetOldest
//
//	consumer := ShipmentRejectedBySellerConsumer{
//		ready: make(chan bool, 0),
//	}
//
//	ctx, cancel := context.WithCancel(context.Background())
//	client, err := sarama.NewConsumerGroup(brokers, main.App.Config.Kafka.ConsumerGroup, config)
//	if err != nil {
//		log.Panicf("Error creating ShipmentRejectedBySeller consumer group client: %v", err)
//	}
//
//	wg := &sync.WaitGroup{}
//	go func() {
//		wg.Add(1)
//		defer wg.Done()
//		for {
//			if err := client.Consume(ctx, strings.Split(topics, ","), &consumer); err != nil {
//				log.Panicf("Error from ShipmentRejectedBySeller consumer: %v", err)
//			}
//			if ctx.Err() != nil {
//				return
//			}
//			consumer.ready = make(chan bool, 0)
//		}
//	}()
//
//	<-consumer.ready // Await till the consumer has been set up
//	log.Println("ShipmentRejectedBySeller consumer up and running!...")
//
//	sigterm := make(chan os.Signal, 1)
//	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
//	select {
//	case <-ctx.Done():
//		log.Println("terminating: context cancelled")
//	case <-sigterm:
//		log.Println("terminating: via signal")
//	}
//	cancel()
//	wg.Wait()
//	if err = client.Close(); err != nil {
//		log.Panicf("Error closing client: %v", err)
//	}
//}
//
//// Consumer represents a Sarama consumer group consumer
//type ShipmentRejectedBySellerConsumer struct {
//	ready chan bool
//}
//
//// Setup is run at the beginning of a new session, before ConsumeClaim
//func (c *ShipmentRejectedBySellerConsumer) Setup(sarama.ConsumerGroupSession) error {
//	close(c.ready)
//	return nil
//}
//
//// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
//func (c *ShipmentRejectedBySellerConsumer) Cleanup(sarama.ConsumerGroupSession) error {
//	return nil
//}
