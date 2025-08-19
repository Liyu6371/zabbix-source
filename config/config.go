package config

import "github.com/mitchellh/mapstructure"

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
	DBType   string `yaml:"DBType"`
	UserName string `yaml:"UserName"`
	Password string `yaml:"Password"`
	Host     string `yaml:"Host"`
	Port     int    `yaml:"Port"`
	DbName   string `yaml:"DbName"`
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
