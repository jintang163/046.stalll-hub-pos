package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig      `mapstructure:"server"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Redis       RedisConfig       `mapstructure:"redis"`
	NSQ         NSQConfig         `mapstructure:"nsq"`
	MinIO       MinIOConfig       `mapstructure:"minio"`
	JWT         JWTConfig         `mapstructure:"jwt"`
	Wechat      WechatConfig      `mapstructure:"wechat"`
	Alipay      AlipayConfig      `mapstructure:"alipay"`
	ClickHouse  ClickHouseConfig  `mapstructure:"clickhouse"`
	DingTalk    DingTalkConfig    `mapstructure:"dingtalk"`
	Inventory   InventoryConfig   `mapstructure:"inventory"`
	CostAlert   CostAlertConfig   `mapstructure:"cost_alert"`
	Amap        AmapConfig        `mapstructure:"amap"`
	Meituan     MeituanConfig     `mapstructure:"meituan"`
	Eleme       ElemeConfig       `mapstructure:"eleme"`
}

type AmapConfig struct {
	Key      string `mapstructure:"key"`
	WebKey   string `mapstructure:"web_key"`
	BaseURL  string `mapstructure:"base_url"`
}

type MeituanConfig struct {
	AppKey    string `mapstructure:"app_key"`
	AppSecret string `mapstructure:"app_secret"`
	BaseURL   string `mapstructure:"base_url"`
	Enabled   bool   `mapstructure:"enabled"`
}

type ElemeConfig struct {
	AppKey    string `mapstructure:"app_key"`
	AppSecret string `mapstructure:"app_secret"`
	BaseURL   string `mapstructure:"base_url"`
	Enabled   bool   `mapstructure:"enabled"`
}

type ClickHouseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type DingTalkConfig struct {
	Webhook string `mapstructure:"webhook"`
	Secret  string `mapstructure:"secret"`
}

type InventoryConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	BaseURL        string `mapstructure:"base_url"`
	APIKey         string `mapstructure:"api_key"`
	APISecret      string `mapstructure:"api_secret"`
	SyncInterval   int    `mapstructure:"sync_interval"`
	TimeoutSeconds int    `mapstructure:"timeout_seconds"`
}

type CostAlertConfig struct {
	Enabled              bool    `mapstructure:"enabled"`
	PriceChangeThreshold float64 `mapstructure:"price_change_threshold"`
	CooldownHours        int     `mapstructure:"cooldown_hours"`
	OperatingExpenseRate float64 `mapstructure:"operating_expense_rate"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbname"`
	Charset      string `mapstructure:"charset"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type NSQConfig struct {
	LookupdAddress string `mapstructure:"lookupd_address"`
	NSQDAddress    string `mapstructure:"nsqd_address"`
}

type MinIOConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	UseSSL    bool   `mapstructure:"use_ssl"`
	Bucket    string `mapstructure:"bucket"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"secret"`
	ExpireHours int    `mapstructure:"expire_hours"`
}

type WechatConfig struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
	MchID     string `mapstructure:"mch_id"`
	APIKey    string `mapstructure:"api_key"`
}

type AlipayConfig struct {
	AppID      string `mapstructure:"app_id"`
	PrivateKey string `mapstructure:"private_key"`
	PublicKey  string `mapstructure:"public_key"`
	NotifyURL  string `mapstructure:"notify_url"`
	Sandbox    bool   `mapstructure:"sandbox"`
}

var AppConfig *Config

func LoadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: config file not found, using environment variables and defaults: %v", err)
	}

	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	log.Println("Config loaded successfully")
}
