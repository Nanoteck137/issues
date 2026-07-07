package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/nanoteck137/issues"
	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Server string `mapstructure:"server"`
	Token  string `mapstructure:"token"`
}

func configDefaults() map[string]any {
	return map[string]any{
		"server": "https://forgejo.nanoteck137.net",
		"token":  "",
	}
}

func readFileToMap(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var m map[string]any
	err = toml.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func mergeMaps(base, overlay map[string]any) {
	for k, v := range overlay {
		ov, ok := v.(map[string]any)
		if !ok {
			base[k] = v
			continue
		}

		bv, ok := base[k]
		if !ok {
			base[k] = v
			continue
		}

		bvm, ok := bv.(map[string]any)
		if !ok {
			base[k] = v
			continue
		}

		mergeMaps(bvm, ov)
	}
}

func readConfigFromEnv() map[string]any {
	prefix := strings.ToUpper(issues.AppName + "_")
	m := make(map[string]any)

	for _, e := range os.Environ() {
		k, v, ok := strings.Cut(e, "=")
		if !ok {
			continue
		}

		if !strings.HasPrefix(k, prefix) {
			continue
		}

		key := strings.ToLower(strings.TrimPrefix(k, prefix))
		m[key] = v
	}

	return m
}

func DefaultConfigDir() (string, error) {
	userDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(userDir, issues.AppName), nil
}

func Load(cfgFile string) (*Config, error) {
	configMap := configDefaults()

	if cfgFile != "" {
		m, err := readFileToMap(cfgFile)
		if err != nil {
			return nil, fmt.Errorf("reading config: %w", err)
		}

		mergeMaps(configMap, m)
	} else {
		var paths []string

		paths = append(paths, filepath.Join("/etc", issues.AppName, "config.toml"))

		if userDir, err := os.UserConfigDir(); err == nil {
			paths = append(paths, filepath.Join(userDir, issues.AppName, "config.toml"))
		}

		paths = append(paths, "config.toml")

		for _, p := range paths {
			m, err := readFileToMap(p)
			if err != nil {
				if !os.IsNotExist(err) {
					return nil, fmt.Errorf("reading config: %w", err)
				}
				continue
			}

			mergeMaps(configMap, m)
		}
	}

	envMap := readConfigFromEnv()
	mergeMaps(configMap, envMap)

	var config Config

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata:         nil,
		Result:           &config,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return nil, fmt.Errorf("new decoder: %w", err)
	}

	err = decoder.Decode(configMap)
	if err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	return &config, nil
}
