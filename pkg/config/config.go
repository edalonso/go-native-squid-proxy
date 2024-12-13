package config

import (
    "fmt"
    "github.com/spf13/viper"
)

type Config struct {
    MetricServerAddress  string
    ServerAddress         string
    MaxConnections        int
    LogLevel              string
    // Add more configuration fields as needed
}

func LoadConfig() (*Config, error) {
    viper.SetConfigName("config")
    viper.AddConfigPath(".")
    viper.SetDefault("MetricServerAddress", ":8081")
    viper.SetDefault("ServerAddress", ":8080")
    viper.SetDefault("MaxConnections", 10000)
    viper.SetDefault("LogLevel", "info")

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }

    // Validate the configuration
    if config.MetricServerAddress == "" {
        return nil, fmt.Errorf("MetricServerAddress cannot be empty")
    }
    if config.ServerAddress == "" {
        return nil, fmt.Errorf("ServerAddress cannot be empty")
    }
    if config.MaxConnections <= 0 {
        return nil, fmt.Errorf("MaxConnections must be greater than 0")
    }
    return &config, nil
}
