package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// Config はアプリケーション全体の設定を保持します
type Config struct {
	Server ServerConfig
	Auth   AuthConfig
	Log    LogConfig
}

// ServerConfig はHTTPサーバーの設定を保持します
type ServerConfig struct {
	Port            string        // サーバーのポート番号
	ReadTimeout     time.Duration // リクエスト読み込みタイムアウト
	WriteTimeout    time.Duration // レスポンス書き込みタイムアウト
	IdleTimeout     time.Duration // アイドル接続のタイムアウト
	ShutdownTimeout time.Duration // グレースフルシャットダウンのタイムアウト
	MaxHeaderBytes  int           // 最大ヘッダーサイズ
}

// AuthConfig は認証の設定を保持します
type AuthConfig struct {
	SessionTimeout   time.Duration // セッションタイムアウト
	MaxLoginAttempts int           // 最大ログイン試行回数
	LockoutDuration  time.Duration // アカウントロックアウト期間
}

// LogConfig はログの設定を保持します
type LogConfig struct {
	Level  string // ログレベル (debug, info, warn, error)
	Format string // ログフォーマット (json, text)
}

// Load は環境変数から設定を読み込みます
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:            getEnv("SERVER_PORT", "8080"),
			ReadTimeout:     getDurationEnv("SERVER_READ_TIMEOUT", 15*time.Second),
			WriteTimeout:    getDurationEnv("SERVER_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:     getDurationEnv("SERVER_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: getDurationEnv("SERVER_SHUTDOWN_TIMEOUT", 30*time.Second),
			MaxHeaderBytes:  getIntEnv("SERVER_MAX_HEADER_BYTES", 1<<20), // 1MB
		},
		Auth: AuthConfig{
			SessionTimeout:   getDurationEnv("AUTH_SESSION_TIMEOUT", 24*time.Hour),
			MaxLoginAttempts: getIntEnv("AUTH_MAX_LOGIN_ATTEMPTS", 5),
			LockoutDuration:  getDurationEnv("AUTH_LOCKOUT_DURATION", 30*time.Minute),
		},
		Log: LogConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},
	}
}

// getEnv は環境変数を取得し、存在しない場合はデフォルト値を返します
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnv は環境変数を整数として取得し、存在しない場合はデフォルト値を返します
func getIntEnv(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("警告: 環境変数 %s の値が不正です: %v. デフォルト値 %d を使用します", key, err, defaultValue)
		return defaultValue
	}

	return value
}

// getDurationEnv は環境変数を時間として取得し、存在しない場合はデフォルト値を返します
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		log.Printf("警告: 環境変数 %s の値が不正です: %v. デフォルト値 %v を使用します", key, err, defaultValue)
		return defaultValue
	}

	return value
}

// Validate は設定の妥当性を検証します
func (c *Config) Validate() error {
	// ポート番号の検証
	port, err := strconv.Atoi(c.Server.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("無効なポート番号: %s", c.Server.Port)
	}

	// タイムアウト値の検証
	if c.Server.ReadTimeout <= 0 {
		log.Printf("警告: ReadTimeoutが0以下です")
	}
	if c.Server.WriteTimeout <= 0 {
		log.Printf("警告: WriteTimeoutが0以下です")
	}

	// ログレベルの検証
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[c.Log.Level] {
		log.Printf("警告: 無効なログレベル: %s", c.Log.Level)
	}

	return nil
}
