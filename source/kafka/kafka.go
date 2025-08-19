package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"zabbix-source/config"
	"zabbix-source/logger"
	"zabbix-source/source"

	"github.com/IBM/sarama"
)

var (
	//rebalance         = sarama.BalanceStrategyRange
	consumerGroup     = "kafka_default_consumer_group"
	kafkaRebalanceMap = map[string]sarama.BalanceStrategy{
		"sticky":     sarama.BalanceStrategySticky,
		"roundrobin": sarama.BalanceStrategyRoundRobin,
		"range":      sarama.BalanceStrategyRange,
	}
)

type KafkaConfig struct {
	Addr          []string `mapstructure:"addr"`
	Username      string   `mapstructure:"username"`
	Password      string   `mapstructure:"password"`
	Version       string   `mapstructure:"version"`
	ConsumerGroup string   `mapstructure:"kafka_consumer_group"`
	ConsumeOldest bool     `mapstructure:"kafka_oldest"`
	Assignor      string   `mapstructure:"kafka_assignor"`
	Topics        []string `mapstructure:"topics"`
	Worker        int      `mapstructure:"worker"`
}

type KafkaSource struct {
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	conf    KafkaConfig
	group   sarama.ConsumerGroup
	handler sarama.ConsumerGroupHandler
}

func init() {
	if err := source.RegisterSource("kafka", NewKafkaSource); err != nil {
		fmt.Printf("failed to register kafka source: %v\n", err)
	}
}

func NewKafkaSource(conf config.SourceConfig) source.SourceInstance {
	c := KafkaConfig{}
	if err := conf.To(&c); err != nil {
		logger.Errorf("failed to decode kafka config: %v", err)
		return nil
	}
	saramaConf := sarama.NewConfig()
	if c.Version != "" {
		if v, err := sarama.ParseKafkaVersion(c.Version); err == nil {
			saramaConf.Version = v
		}
	}
	if c.Username != "" && c.Password != "" {
		saramaConf.Net.SASL.Enable = true
		saramaConf.Net.SASL.User = c.Username
		saramaConf.Net.SASL.Password = c.Password
	}
	if c.ConsumeOldest {
		saramaConf.Consumer.Offsets.Initial = sarama.OffsetOldest
	} else {
		saramaConf.Consumer.Offsets.Initial = sarama.OffsetNewest
	}

	if v, ok := kafkaRebalanceMap[c.Assignor]; ok {
		saramaConf.Consumer.Group.Rebalance.Strategy = v
	} else {
		saramaConf.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	}

	if c.ConsumerGroup != "" {
		consumerGroup = c.ConsumerGroup
	}
	group, err := sarama.NewConsumerGroup(c.Addr, consumerGroup, saramaConf)
	if err != nil {
		logger.Errorf("failed to create kafka consumer group: %v", err)
		return nil
	}
	return &KafkaSource{
		wg:    sync.WaitGroup{},
		conf:  c,
		group: group,
	}
}

func (k *KafkaSource) Name() string {
	return "kafka"
}

func (k *KafkaSource) Run(ch chan<- []byte) error {
	k.ctx, k.cancel = context.WithCancel(context.Background())
	k.handler = &Handler{ch: ch}

	worker := 3
	if k.conf.Worker > 0 {
		worker = k.conf.Worker
	}
	k.wg.Add(worker)
	for idx := 0; idx < worker; idx++ {
		go func(idx int) {
			defer k.wg.Done()
			for {
				if err := k.group.Consume(k.ctx, k.conf.Topics, k.handler); err != nil {
					if errors.Is(err, sarama.ErrClosedConsumerGroup) {
						logger.Errorf("kafka consumer goroutine %d group closed: %v", idx, err)
						return
					}
					logger.Errorf("kafka consumer goroutine %d failed to consume from kafka: %v", idx, err)
				}
				if k.ctx.Err() != nil {
					logger.Infof("kafka consumer goroutine %d context cancelled", idx)
					return
				}
			}
		}(idx)
	}
	return nil
}

func (k *KafkaSource) Stop() {
	k.cancel()
	k.wg.Wait()
	if err := k.group.Close(); err != nil {
		logger.Errorf("failed to close kafka consumer group: %v", err)
	}
}

type Handler struct {
	ch chan<- []byte
}

func (h *Handler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				logger.Infof("Source->kafka message chan claim is closed.")
				return nil
			}
			h.ch <- msg.Value
		case <-session.Context().Done():
			logger.Info("session exit")
			return nil
		}
	}
}
