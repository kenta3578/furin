package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// デフォルトのサウンド設定
const (
	defaultSoundName   = "Glass"
	defaultSoundVolume = 1.0
	minSoundVolume     = 0.0
	maxSoundVolume     = 2.0
)

// Config はアプリケーション設定全体を表す
type Config struct {
	Timers       []TimerConfig `yaml:"timers"`
	SoundEnabled *bool         `yaml:"sound_enabled,omitempty"` // nil の場合はデフォルト true
	SoundName    string        `yaml:"sound_name,omitempty"`    // 空の場合はデフォルト Glass
	SoundVolume  *float64      `yaml:"sound_volume,omitempty"`  // nil の場合はデフォルト 1.0
}

// TimerConfig は YAML に保存するタイマー設定を表す
type TimerConfig struct {
	Name   string `yaml:"name"`
	Emoji  string `yaml:"emoji"`
	Action string `yaml:"action,omitempty"` // 完了時にすべきアクション（任意）
	// Type は "interval" / "once"。省略時は interval として扱う（後方互換）。
	Type     string `yaml:"type,omitempty"`
	Interval int    `yaml:"interval,omitempty"` // 分単位（interval 用）
	// At は once タイマーの発火時刻 "HH:MM"（once 用）
	At string `yaml:"at,omitempty"`
	// Fired は once タイマーが発火済みかどうか（「済み」状態の永続化用）
	Fired bool `yaml:"fired,omitempty"`
}

// boolPtr は bool 値へのポインタを返すヘルパー
func boolPtr(b bool) *bool { return &b }

// defaultConfig はデフォルト設定を返す
func defaultConfig() *Config {
	return &Config{
		Timers: []TimerConfig{
			{Name: "水分補給", Emoji: "💧", Interval: 30},
			{Name: "座りすぎ防止", Emoji: "🪑", Interval: 60},
		},
		SoundEnabled: boolPtr(true),
		SoundName:    defaultSoundName,
		SoundVolume:  float64Ptr(defaultSoundVolume),
	}
}

// float64Ptr は float64 値へのポインタを返すヘルパー
func float64Ptr(f float64) *float64 { return &f }

// IsSoundEnabled は SoundEnabled の値を返す。nil の場合は true を返す
func (c *Config) IsSoundEnabled() bool {
	if c.SoundEnabled == nil {
		return true
	}
	return *c.SoundEnabled
}

// GetSoundName は SoundName の値を返す。空の場合はデフォルト Glass を返す
func (c *Config) GetSoundName() string {
	if c.SoundName == "" {
		return defaultSoundName
	}
	return c.SoundName
}

// GetSoundVolume は SoundVolume の値を返す。nil の場合はデフォルト 1.0 を返す。
// 範囲外の値は 0.0〜2.0 にクランプする
func (c *Config) GetSoundVolume() float64 {
	if c.SoundVolume == nil {
		return defaultSoundVolume
	}
	return clampVolume(*c.SoundVolume)
}

// clampVolume はボリュームを 0.0〜2.0 の範囲に収める
func clampVolume(v float64) float64 {
	if v < minSoundVolume {
		return minSoundVolume
	}
	if v > maxSoundVolume {
		return maxSoundVolume
	}
	return v
}

// ConfigPath は設定ファイルのデフォルトパスを返す。
// XDG_CONFIG_HOME が設定されていればそちらを優先し、
// 未設定の場合は ~/.config/furin/config.yaml を返す。
func ConfigPath() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("ホームディレクトリの取得に失敗: %w", err)
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "furin", "config.yaml"), nil
}

// LoadConfig は指定パスから設定を読み込む。
// ファイルが存在しない場合はデフォルト設定を返す。
// ファイルが破損・不正フォーマットの場合はエラーを返す。
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaultConfig(), nil
		}
		return nil, fmt.Errorf("設定ファイルの読み込みに失敗: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("設定ファイルの解析に失敗: %w", err)
	}

	return &cfg, nil
}

// SaveConfig は設定を指定パスに書き込む。
// 必要に応じてディレクトリを作成する。
func SaveConfig(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("設定ディレクトリの作成に失敗: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("設定のシリアライズに失敗: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("設定ファイルの書き込みに失敗: %w", err)
	}

	return nil
}

// timersFromConfig は Config の TimerConfig スライスを Timer スライスに変換する。
// type が省略された既存 YAML は interval として扱う（後方互換）。
func timersFromConfig(cfg *Config) []Timer {
	now := time.Now()
	timers := make([]Timer, 0, len(cfg.Timers))
	for _, tc := range cfg.Timers {
		if tc.Type == TimerTypeOnce {
			hour, min, ok := parseHHMM(tc.At)
			t := Timer{
				Name:      tc.Name,
				Emoji:     tc.Emoji,
				Action:    tc.Action,
				Type:      TimerTypeOnce,
				Notifying: tc.Fired,
				Fired:     tc.Fired,
			}
			if ok {
				t.FireAt = fireAtToday(hour, min, now)
			}
			timers = append(timers, t)
			continue
		}

		d := time.Duration(tc.Interval) * time.Minute
		timers = append(timers, Timer{
			Name:      tc.Name,
			Emoji:     tc.Emoji,
			Action:    tc.Action,
			Type:      TimerTypeInterval,
			Interval:  d,
			Remaining: d,
		})
	}
	return timers
}

// timersToConfig は Timer スライスを Config の TimerConfig スライスに変換する
func timersToConfig(timers []Timer) *Config {
	cfg := &Config{
		Timers: make([]TimerConfig, 0, len(timers)),
	}
	for _, t := range timers {
		if t.IsOnce() {
			cfg.Timers = append(cfg.Timers, TimerConfig{
				Name:   t.Name,
				Emoji:  t.Emoji,
				Action: t.Action,
				Type:   TimerTypeOnce,
				At:     formatHHMM(t.FireAt.Hour(), t.FireAt.Minute()),
				Fired:  t.Fired,
			})
			continue
		}

		cfg.Timers = append(cfg.Timers, TimerConfig{
			Name:     t.Name,
			Emoji:    t.Emoji,
			Action:   t.Action,
			Type:     TimerTypeInterval,
			Interval: int(t.Interval.Minutes()),
		})
	}
	return cfg
}
