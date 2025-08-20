package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
)

var (
	globalConfig *Config
)

type SenderConfig map[string]any

func (s SenderConfig) To(out interface{}) error {
	return mapstructure.Decode(s, out)
}

type SourceConfig map[string]any

func (s SourceConfig) To(out interface{}) error {
	return mapstructure.Decode(s, out)
}

// SQLConfig 数据库配置
type SQLConfig struct {
	// 数据库类型 一般情况下常用的类型为 MySQL PostgreSQL SQLite
	DBType   string `yaml:"db_type"`
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	DbName   string `yaml:"db_name"`
}

type ZabbixConfig struct {
	SQLConfig SQLConfig `yaml:"sql_config"`
}

type LoggerConfig struct {
	Level      string `yaml:"level"`
	OutputPath string `yaml:"output_path"`
}

type Config struct {
	PidFilePath  string                  `yaml:"pid_file_path"`
	ZabbixConfig ZabbixConfig            `yaml:"zabbix_config"`
	LoggerConfig LoggerConfig            `yaml:"logger_config"`
	SenderConfig map[string]SenderConfig `yaml:"sender_config"`
	SourceConfig map[string]SourceConfig `yaml:"source_config"`
}

// Parse 解析配置文件
func Parse(path string) (*Config, error) {
	if path == "" {
		return nil, fmt.Errorf("config file path is empty")
	}
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %v", err)
	}
	cPath := filepath.Join(dir, path)
	content, err := os.ReadFile(cPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %v", cPath, err)
	}
	if err := yaml.Unmarshal(content, &globalConfig); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config file %s: %v", cPath, err)
	}
	return globalConfig, nil
}

func GetGlobalConfig() *Config {
	return globalConfig
}
