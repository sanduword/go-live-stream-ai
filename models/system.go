package models

import "time"

// 系统配置
type Config struct {
	Server ServerConfig
	Db     DbConfig
	Redis  RedisConfig
}

// 服务器配置
type ServerConfig struct {
	Mode     string
	Port     int
	MaxAi    int
	LogPath  string
	YoloPath string
	PathImg  string
	HttpImg  string
	Score    float32
	Nms      float32
}

// 数据库配置
type DbConfig struct {
	Sqlite DbType
}

// 数据库类型
type DbType struct {
	Conn string
}

// Redis
type RedisConfig struct {
	Addr         string
	RdPrefix     string
	RdExpire     int
	Db           int
	MaxIdle      int
	MaxActive    int
	DialTimeout  time.Duration
	IdleTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}
