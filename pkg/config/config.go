package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds Spotify API credentials
type Config struct {
	SpotifyClientID     string `mapstructure:"spotify_client_id"`
	SpotifyClientSecret string `mapstructure:"spotify_client_secret"`
}

// InitConfig sets up configuration directory and default values
func InitConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Create config directory in user's home/.config/mufetch
	configDir := filepath.Join(home, ".config", "mufetch")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	viper.AddConfigPath(configDir)
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Set default empty values for credentials
	viper.SetDefault("spotify_client_id", "")
	viper.SetDefault("spotify_client_secret", "")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return viper.SafeWriteConfig() // Create config file if not found
		}
		return err
	}

	return nil
}

// GetConfig unmarshals configuration into Config struct
func GetConfig() (*Config, error) {
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

// SetCredentials saves Spotify API credentials to config file
func SetCredentials(clientID, clientSecret string) error {
	viper.Set("spotify_client_id", clientID)
	viper.Set("spotify_client_secret", clientSecret)
	return viper.WriteConfig()
}

// HasCredentials checks if valid Spotify credentials are configured
func HasCredentials() bool {
	config, err := GetConfig()
	if err != nil {
		return false
	}
	return config.SpotifyClientID != "" && config.SpotifyClientSecret != ""
}

// init initializes the viper configuration
func init() {
	viper.AutomaticEnv()
	viper.SetEnvPrefix("MUFETCH") // MUFETCH_SPOTIFY_CLIENT_ID, MUFETCH_SPOTIFY_CLIENT_SECRET
}
