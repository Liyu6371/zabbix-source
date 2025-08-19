package source

import (
	"fmt"
	"strings"
	"zabbix-source/config"
)

type SourceInstance interface {
	// Name 返回 Source 实例的名称
	Name() string
	// Run 启动 Source 实例
	Run(chan<- []byte) error
	// Stop 停止 Source 实例
	Stop()
}

var sourceFactory = make(map[string]func(config.SourceConfig) SourceInstance)

func RegisterSource(name string, factory func(config.SourceConfig) SourceInstance) error {
	_, ok := sourceFactory[name]
	if ok {
		return fmt.Errorf("source %s already registered", name)
	}
	sourceFactory[name] = factory
	return nil
}

type SourceService struct {
	ch        chan []byte
	instances map[string]SourceInstance
	conf      map[string]config.SourceConfig
}

func NewSourceService(conf map[string]config.SourceConfig) (*SourceService, error) {
	if len(conf) == 0 {
		return nil, fmt.Errorf("no source configurations provided")
	}
	return &SourceService{
		ch:        make(chan []byte, 500),
		instances: make(map[string]SourceInstance),
		conf:      conf,
	}, nil
}

func (s *SourceService) Start() error {
	var errArray []error
	for name, cfg := range s.conf {
		factory, ok := sourceFactory[name]
		if !ok {
			errArray = append(errArray, fmt.Errorf("source %s not registered", name))
			continue
		}
		instance := factory(cfg)
		if instance == nil {
			errArray = append(errArray, fmt.Errorf("failed to create instance for source %s", name))
			continue
		}
		if err := instance.Run(s.ch); err != nil {
			errArray = append(errArray, fmt.Errorf("failed to run source %s: %v", name, err))
			continue
		}
		s.instances[name] = instance
	}
	var errMsgs []string
	for _, err := range errArray {
		errMsgs = append(errMsgs, err.Error())
	}
	if len(errMsgs) > 0 {
		return fmt.Errorf("errors occurred while starting source: %s", strings.Join(errMsgs, "\n"))
	}
	if len(s.instances) == 0 {
		return fmt.Errorf("no sources started successfully")
	}
	return nil
}

func (s *SourceService) Stop() {
	// 关闭所有的 生产者
	for _, instance := range s.instances {
		instance.Stop()
	}
	close(s.ch)
}

func (s *SourceService) Chan() <-chan []byte {
	return s.ch
}
