package kafkaadapter

import (
	"errors"
	"time"

	"github.com/Shopify/sarama"
)

func CreateTopic(broker, topic string, partitions int32, replication int16) error {
	// Set broker configuration
	b := sarama.NewBroker(broker)

	// Additional configurations. Check sarama doc for more info
	config := sarama.NewConfig()
	config.Version = sarama.V1_0_0_0

	// Open broker connection with configs defined above
	err := b.Open(config)
	if err != nil {
		return err
	}

	// check if the connection was OK
	connected, err := b.Connected()
	if err != nil {
		return err
	}

	if !connected {
		return errors.New("cant connect to broker")
	}

	// Setup the Topic details in CreateTopicRequest struct
	t := topic
	topicDetail := &sarama.TopicDetail{}
	topicDetail.NumPartitions = partitions
	topicDetail.ReplicationFactor = replication
	topicDetail.ConfigEntries = make(map[string]*string)

	topicDetails := make(map[string]*sarama.TopicDetail)
	topicDetails[t] = topicDetail

	request := sarama.CreateTopicsRequest{
		Timeout:      time.Second * 15,
		TopicDetails: topicDetails,
	}

	// Send request to Broker
	res, err := b.CreateTopics(&request)

	// handle errors if any
	if err != nil {
		return err
	}

	te := res.TopicErrors
	for _, v := range te {
		if v.Err != sarama.ErrNoError {
			return v.Err
		}
	}

	// close connection to broker
	err = b.Close()
	if err != nil {
		return err
	}
	return nil
}

func GetTopics(brokers []string) ([]string, error) {
	conf := sarama.NewConfig()
	conf.Consumer.Return.Errors = true
	consumer, err := sarama.NewConsumer(brokers, conf)
	if err != nil {
		return nil, err
	}
	topics, err := consumer.Topics()
	if err != nil {
		return nil, err
	}

	return topics, nil
}

func (k *Kafka) CreatePartition(topic string, partition int32) error {
	k.Config.Version = sarama.V2_2_0_0
	cAdmin, err := sarama.NewClusterAdmin(k.brokers, k.Config)
	if err != nil {
		return err
	}
	return cAdmin.CreatePartitions(topic, partition, nil, false)
}
