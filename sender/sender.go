package sender

import (
	"fmt"
	"strings"
	"sync"
	"zabbix-source/config"
	"zabbix-source/logger"
)

// SenderMsg 消息接口
// 负责提供消息数据和补充信息
// 补充信息用于具体的Sender实现
type SenderMsg interface {
	// GetData 返回消息数据
	GetData() []byte
	// GetOptions 返回消息的补充信息
	GetOptions() map[string]interface{}
	// GetSender 返回消息的发送者
	GetSender() string
}

// SenderInstance Sender 实例接口
type SenderInstance interface {
	// Name 返回 Sender 实例的名称
	Name() string
	// Run 启动 Sender 实例
	Run() error
	// Push 发送消息
	Push(SenderMsg)
	// Stop 停止 Sender 实例
	Stop()
}

var senderFactory = make(map[string]func(config.SenderConfig) SenderInstance)

func RegisterSender(name string, factory func(config.SenderConfig) SenderInstance) error {
	_, ok := senderFactory[name]
	if ok {
		return fmt.Errorf("sender %s already registered", name)
	}
	senderFactory[name] = factory
	return nil
}

type SenderService struct {
	wg        sync.WaitGroup
	conf      map[string]config.SenderConfig
	msgChan   chan SenderMsg
	instances map[string]SenderInstance
}

func NewSenderService(conf map[string]config.SenderConfig) (*SenderService, error) {
	if len(conf) == 0 {
		return nil, fmt.Errorf("no sender configurations provided")
	}
	return &SenderService{
		wg:        sync.WaitGroup{},
		conf:      conf,
		msgChan:   make(chan SenderMsg, 500),
		instances: make(map[string]SenderInstance),
	}, nil
}

// Start 启动 SenderService
// error 返回给上层进行处理如若出现 error 不为 nil
// defer 调用 Stop 方法停止服务
func (s *SenderService) Start() error {
	var errArray []error
	for name, cfg := range s.conf {
		factory, ok := senderFactory[name]
		if !ok {
			errArray = append(errArray, fmt.Errorf("sender %s not registered", name))
			continue
		}
		sender := factory(cfg)
		if sender == nil {
			errArray = append(errArray, fmt.Errorf("failed to create sender %s", name))
			continue
		}
		if err := sender.Run(); err != nil {
			errArray = append(errArray, fmt.Errorf("failed to run sender %s: %v", sender.Name(), err))
			continue
		}
		s.instances[name] = sender
	}
	var errMsgs []string
	for _, err := range errArray {
		errMsgs = append(errMsgs, err.Error())
	}
	if len(errMsgs) > 0 {
		return fmt.Errorf("errors occurred while starting senders: %s", strings.Join(errMsgs, "\n"))
	}
	// 启动消息分发 goroutine
	s.wg.Add(3)
	for idx := 0; idx < 3; idx++ {
		go s.dispatch(idx)
	}

	return nil
}

// dispatch SenderService 消息分发
func (s *SenderService) dispatch(index int) {
	defer s.wg.Done()
	for msg := range s.msgChan {
		name := msg.GetSender()
		sender, ok := s.instances[name]
		if !ok {
			logger.Errorf("dispatch to sender %s, instance not found", name)
			continue
		}
		sender.Push(msg)
	}
	logger.Infof("dispatch goroutine %d exit", index)
}

func (s *SenderService) Stop() {
	for _, sender := range s.instances {
		sender.Stop()
	}
	close(s.msgChan)
	s.wg.Wait()
}
