package source

import (
	"fmt"
	"sync"
	"zabbix-source/config"
)

type SourceInstance interface {
	// Name 返回 Source 实例的名称
	Name() string
	// Run 启动 Source 实例
	Run(ch []byte) error
	// Stop 停止 Source 实例
	Stop()
}

var sourceFactory = make(map[string]func() SourceInstance)

func RegisterSource(name string, factory func() SourceInstance) error {
	_, ok := sourceFactory[name]
	if ok {
		return fmt.Errorf("source %s already registered", name)
	}
	sourceFactory[name] = factory
	return nil
}

type SourceService struct {
	ch        chan []byte
	wg        sync.WaitGroup
	instances map[string]SourceInstance
	conf      map[string]config.SourceConfig
}

func NewSourceService(conf map[string]config.SourceConfig) (*SourceService, error) {
	if len(conf) == 0 {
		return nil, fmt.Errorf("no source configurations provided")
	}
	return &SourceService{
		ch:        make(chan []byte, 500),
		wg:        sync.WaitGroup{},
		instances: make(map[string]SourceInstance),
		conf:      conf,
	}, nil
}

func (s *SourceService) Start() error {
	return nil
}

func (s *SourceService) Stop() {}
