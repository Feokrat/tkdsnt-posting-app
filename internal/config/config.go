package config

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"

	"github.com/spf13/viper"
)

type (
	Config struct {
		Postgresql PGConfig
		Posting    Posting
	}

	PGConfig struct {
		Host     string `mapstructure:"host"`
		Port     string `mapstructure:"port"`
		Username string `mapstructure:"user"`
		Password string `mapstructure:"password"`
		DBName   string `mapstructure:"dbname"`
		SSLMode  string `mapstructure:"ssl"`
	}

	Posting struct {
		AccessToken string `json:"access_token"`
		GroupId     int    `json:"group_id"`
		GelbooruAccessKey string `json:"gelbooru_access_key"`
		GelbooruUserId int `json:"gelbooru_user_id"`
	}
)

func Init(path string, logger *log.Logger) (*Config, error) {
	if err := parseConfigFile(path); err != nil {
		logger.Printf("failed to parse path to config file: %s", err)
		return nil, err
	}

	var cfg Config
	if err := unmarshal(&cfg, logger); err != nil {
		logger.Printf("failed to unmarshal config: %s", err)
		return nil, err
	}

	setFromEnv(&cfg)

	return &cfg, nil
}

func setFromEnv(cfg *Config) {
	godotenv.Load(".env")
	cfg.Posting.AccessToken = os.Getenv("ACCESS_TOKEN")
	cfg.Posting.GroupId, _ = strconv.Atoi(os.Getenv("GROUP_ID"))
	cfg.Posting.GelbooruAccessKey = os.Getenv("GELBOORU_ACCESS_KEY")
	cfg.Posting.GelbooruUserId, _ = strconv.Atoi(os.Getenv("GELBOORU_USER_ID"))
}

func unmarshal(cfg *Config, logger *log.Logger) error {
	if err := viper.UnmarshalKey("postgres", &cfg.Postgresql); err != nil {
		logger.Printf("failed to unmarshal postgres key in config: %s", err)
		return err
	}

	return nil
}

func parseConfigFile(filepath string) error {
	path := strings.Split(filepath, "/")

	viper.AddConfigPath(path[0])
	viper.SetConfigName(path[1])

	return viper.ReadInConfig()
}
