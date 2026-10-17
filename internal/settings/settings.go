package settings

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v2"
)

type BotSettings struct {
	Discord DiscordSettings `yaml:"discord"`
}

type DiscordSettings struct {
	Token   string `yaml:"token"`
	GuildID string `yaml:"gid"`
}

func ReadEnvironmentVariable(
	value string,
) (string, error) {
	envValue := os.Getenv(value)
	if len(envValue) == 0 {
		return "", errors.New(
			"environment variable " + value + " doesn't exist",
		)
	}

	return envValue, nil
}

func GetDefaultConfigPath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Join(
		filepath.Dir(executable),
		"config.yaml",
	), nil
}

func LoadConfig(
	path string,
) (*BotSettings, error) {
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	settings := &BotSettings{}
	if err := yaml.Unmarshal(configBytes, settings); err != nil {
		return nil, err
	}

	if strings.TrimSpace(settings.Discord.Token) == "" {
		return nil, errors.New("discord token is empty")
	}

	return settings, nil
}

func LoadEnvironmentConfig(
	path string,
) (*BotSettings, error) {
	settings, err := LoadConfig(path)
	if err != nil {
		return nil, err
	}

	settings.Discord.Token, err = ReadEnvironmentVariable(settings.Discord.Token)
	if err != nil {
		return nil, err
	}

	if settings.Discord.GuildID != "" {
		settings.Discord.GuildID, err = ReadEnvironmentVariable(settings.Discord.GuildID)
		if err != nil {
			return nil, err
		}
	}

	return settings, nil
}

func SaveConfig(
	settings *BotSettings,
	path string,
) error {
	if settings == nil {
		return errors.New("settings are nil")
	}

	if strings.TrimSpace(settings.Discord.Token) == "" {
		return errors.New("discord token is empty")
	}

	configBytes, err := yaml.Marshal(settings)
	if err != nil {
		return err
	}

	return os.WriteFile(path, configBytes, 0600)
}

func GetConfigPathFromArgs() string {
	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"Usage: %s [path_to_config_file]\n",
			os.Args[0],
		)
		flag.PrintDefaults()
	}

	flag.Parse()
	if path := flag.Arg(0); path != "" {
		return path
	}

	path, err := GetDefaultConfigPath()
	if err != nil {
		return "config.yaml"
	}

	return path
}

func Get() *BotSettings {
	return &BotSettings{}
}
