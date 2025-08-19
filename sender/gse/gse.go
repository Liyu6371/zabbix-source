package gse

import (
	"fmt"
	"sync"
	"time"
	"zabbix-source/config"
	"zabbix-source/logger"
	"zabbix-source/sender"

	"github.com/TencentBlueKing/bkmonitor-datalink/pkg/libgse/gse"
)

var (
	worker         = 3
	buffer         = 500
	defaultGseConf = gse.Config{
		MsgQueueSize:   1,
		WriteTimeout:   5 * time.Second,
		ReadTimeout:    60 * time.Second,
		Nonblock:       false,
		RetryTimes:     3,
		RetryInterval:  3 * time.Second,
		ReconnectTimes: 3,
		Endpoint:       "/var/run/ipc.state.report",
	}
)

func init() {
	// 可能存在 logger 未被初始化的情况
	if err := sender.RegisterSender("gse", NewGseSender); err != nil {
		fmt.Println(err)
	}
}

type GseConfig struct {
	Worker   int    `mapstructure:"worker"`
	Buffer   int    `mapstructure:"buffer"`
	EndPoint string `mapstructure:"end_point"`
}

type GseSender struct {
	cfg    GseConfig
	wg     sync.WaitGroup
	ch     chan sender.SenderMsg
	client *gse.GseClient
}

func NewGseSender(cfg config.SenderConfig) sender.SenderInstance {
	c := GseConfig{}
	if err := cfg.To(&c); err != nil {
		logger.Errorf("failed to decode GSE config: %v", err)
		return nil
	}
	if c.EndPoint != "" {
		defaultGseConf.Endpoint = c.EndPoint
	}
	client, err := gse.NewGseClientFromConfig(defaultGseConf)
	if err != nil {
		logger.Errorf("failed to create GSE client: %v", err)
		return nil
	}
	if c.Buffer > 0 {
		buffer = c.Buffer
	}
	if c.Worker > 0 {
		worker = c.Worker
	}
	return &GseSender{
		cfg:    c,
		wg:     sync.WaitGroup{},
		ch:     make(chan sender.SenderMsg, buffer),
		client: client,
	}
}

func (g *GseSender) Name() string {
	return "gse"
}

func (g *GseSender) Run() error {
	if err := g.client.Start(); err != nil {
		return fmt.Errorf("failed to start GSE client: %v", err)
	}
	g.wg.Add(worker)
	for idx := 0; idx < worker; idx++ {
		go g.consume(idx)
	}
	return nil
}

func (g *GseSender) consume(idx int) {
	defer g.wg.Done()
	for msg := range g.ch {
		options := msg.GetOptions()
		dataid, ok := options["dataid"].(int32)
		if !ok {
			logger.Errorf("GSE sender worker %d: missing or invalid dataid", idx)
			logger.Debugf("GSE sender worker %d, drop msg %s ,options: %v", idx, msg.GetData(), options)
			continue
		}
		if err := g.client.Send(gse.NewGseCommonMsg(msg.GetData(), dataid, 0, 0, 0)); err != nil {
			logger.Errorf("GSE sender worker %d: failed to send message: %v", idx, err)
			continue
		}
	}
	logger.Infof("GSE sender worker %d exiting", idx)
}

func (g *GseSender) Push(msg sender.SenderMsg) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("GSE sender push failed: channel closed")
		}
	}()
	g.ch <- msg
}

func (g *GseSender) Stop() {
	close(g.ch)
	g.wg.Wait()
	g.client.Close()
	logger.Infof("GSE sender stopped")
}
