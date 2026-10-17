package settings

import (
	"errors"
	"flag"
	"fmt"
	"os"

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
		return "", errors.New("environment variable " + value + " doesn't exist")
	}

	return envValue, nil
}

func LoadConfig(
	path string,
) (*BotSettings, error) {
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	settings := &BotSettings{}
	err = yaml.Unmarshal(configBytes, settings)
	if err != nil {
		return nil, err
	}

	settings.Discord.Token, err = ReadEnvironmentVariable(settings.Discord.Token)
	if err != nil {
		return nil, err
	}

	settings.Discord.GuildID, err = ReadEnvironmentVariable(settings.Discord.GuildID)
	if err != nil {
		return nil, err
	}

	return settings, nil
}

func GetConfigPathFromArgs() string {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <path_to_config_file>\n", os.Args[0])
		flag.PrintDefaults()
	}

	var configPath string
	flag.Parse()
	if configPath = flag.Arg(0); configPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	return configPath
}

func Get() *BotSettings {
	s := &BotSettings{}
	return s
}
