package kafkaadapter

import (
	"github.com/Shopify/sarama"
)

func CheckBrokers(brokers []string) (bool, error) {
	conf := sarama.NewConfig()
	client, err := sarama.NewClient(brokers, conf)
	if err != nil {
		return false, err
	}
	err = client.Close()
	if err != nil {
		return false, err
	}
	return true, nil
}

func ChangeReplicaFactor(brokers []string, topic, replicaNumber string) error {
	k := NewKafka(brokers, "")
	k.Config.Version = sarama.V2_2_0_0
	admin, err := sarama.NewClusterAdmin(brokers, k.Config)
	if err != nil {
		return err
	}

	var value string
	entries := make(map[string]*string)
	value = replicaNumber
	entries["ReplicationFactor"] = &value
	err = admin.AlterConfig(sarama.TopicResource, topic, entries, false)
	if err != nil {
		return err
	}

	err = admin.Close()
	if err != nil {
		return err
	}
	return nil
}
