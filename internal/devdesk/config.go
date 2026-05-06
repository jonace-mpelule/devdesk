package devdesk

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultConfigFile = ".devdesk.yaml"

type Config struct {
	Presets map[string]Preset `yaml:"presets"`
}

type Preset struct {
	CloseAll bool     `yaml:"close_all,omitempty"`
	Close    []string `yaml:"close,omitempty"`
	Open     []string `yaml:"open,omitempty"`
	Commands []string `yaml:"commands,omitempty"`
}

func DefaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	return filepath.Join(home, defaultConfigFile), nil
}

func ExpandPath(path string) (string, error) {
	if path == "" {
		return DefaultConfigPath()
	}

	if path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		return home, nil
	}

	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolve home directory: %w", err)
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
	}

	return path, nil
}

func LoadConfig(path string) (Config, error) {
	resolved, err := ExpandPath(path)
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(resolved)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, MissingConfigError{Path: resolved}
		}
		return Config{}, fmt.Errorf("read config %s: %w", resolved, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config %s: %w", resolved, err)
	}
	if cfg.Presets == nil {
		cfg.Presets = map[string]Preset{}
	}

	return cfg, nil
}

func SaveConfig(path string, cfg Config) error {
	resolved, err := ExpandPath(path)
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("format config %s: %w", resolved, err)
	}

	if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
		return fmt.Errorf("create config directory %s: %w", filepath.Dir(resolved), err)
	}

	if err := os.WriteFile(resolved, data, 0o600); err != nil {
		return fmt.Errorf("write config %s: %w", resolved, err)
	}

	return nil
}

func LoadConfigForUpdate(path string) (Config, error) {
	cfg, err := LoadConfig(path)
	if err == nil {
		return cfg, nil
	}

	var missing MissingConfigError
	if errors.As(err, &missing) {
		return Config{Presets: map[string]Preset{}}, nil
	}

	return Config{}, err
}

func (c *Config) SetPreset(name string, preset Preset, overwrite bool) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("workspace name is required")
	}
	if err := preset.Validate(); err != nil {
		return err
	}
	if c.Presets == nil {
		c.Presets = map[string]Preset{}
	}
	if _, exists := c.Presets[name]; exists && !overwrite {
		return fmt.Errorf("workspace %q already exists; use --force to replace it", name)
	}

	c.Presets[name] = preset
	return nil
}

func (c Config) Preset(name string) (Preset, error) {
	preset, ok := c.Presets[name]
	if !ok {
		return Preset{}, UnknownPresetError{Name: name, Available: c.PresetNames()}
	}

	if err := preset.Validate(); err != nil {
		return Preset{}, fmt.Errorf("preset %q: %w", name, err)
	}

	return preset, nil
}

func (c Config) PresetNames() []string {
	names := make([]string, 0, len(c.Presets))
	for name := range c.Presets {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (p Preset) Validate() error {
	if len(p.Close) == 0 && len(p.Open) == 0 && len(p.Commands) == 0 && !p.CloseAll {
		return errors.New("must define close_all, close, open, or commands")
	}
	if p.CloseAll && len(p.Close) > 0 {
		return errors.New("close_all cannot be combined with close apps")
	}

	for _, app := range append(append([]string{}, p.Close...), p.Open...) {
		if strings.TrimSpace(app) == "" {
			return errors.New("app names cannot be empty")
		}
	}

	for _, command := range p.Commands {
		if strings.TrimSpace(command) == "" {
			return errors.New("commands cannot be empty")
		}
	}

	return nil
}

type MissingConfigError struct {
	Path string
}

func (e MissingConfigError) Error() string {
	return fmt.Sprintf("config not found at %s\n\nCreate it with a preset like:\n%s", e.Path, ExampleConfig())
}

type UnknownPresetError struct {
	Name      string
	Available []string
}

func (e UnknownPresetError) Error() string {
	if len(e.Available) == 0 {
		return fmt.Sprintf("unknown preset %q; no presets are configured", e.Name)
	}
	return fmt.Sprintf("unknown preset %q; available presets: %s", e.Name, strings.Join(e.Available, ", "))
}

func ExampleConfig() string {
	return `presets:
  main-workspace:
    close_all: false
    close:
      - Example Chat
    open:
      - Example Editor`
}
