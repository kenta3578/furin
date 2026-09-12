package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestLoadConfig_FileNotExist はファイルが存在しない場合にデフォルト設定が返ることを確認する
func TestLoadConfig_FileNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("存在しないファイルでエラーが発生した: %v", err)
	}
	if cfg == nil {
		t.Fatal("設定がnilになっている")
	}
	def := defaultConfig()
	if len(cfg.Timers) != len(def.Timers) {
		t.Fatalf("タイマー数が異なる: got %d, want %d", len(cfg.Timers), len(def.Timers))
	}
	for i, tc := range cfg.Timers {
		want := def.Timers[i]
		if tc.Name != want.Name || tc.Emoji != want.Emoji || tc.Interval != want.Interval {
			t.Errorf("タイマー[%d]が異なる: got %+v, want %+v", i, tc, want)
		}
	}
}

// TestSaveAndLoadConfig は書き込み→読み込みでデータが一致することを確認する
func TestSaveAndLoadConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "furin", "config.yaml")

	original := &Config{
		Timers: []TimerConfig{
			{Name: "テスト1", Emoji: "🔔", Interval: 15},
			{Name: "テスト2", Emoji: "⏰", Interval: 45},
		},
	}

	if err := SaveConfig(path, original); err != nil {
		t.Fatalf("SaveConfig に失敗: %v", err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig に失敗: %v", err)
	}

	if len(loaded.Timers) != len(original.Timers) {
		t.Fatalf("タイマー数が異なる: got %d, want %d", len(loaded.Timers), len(original.Timers))
	}

	for i, tc := range loaded.Timers {
		want := original.Timers[i]
		if tc.Name != want.Name {
			t.Errorf("タイマー[%d].Name が異なる: got %q, want %q", i, tc.Name, want.Name)
		}
		if tc.Emoji != want.Emoji {
			t.Errorf("タイマー[%d].Emoji が異なる: got %q, want %q", i, tc.Emoji, want.Emoji)
		}
		if tc.Interval != want.Interval {
			t.Errorf("タイマー[%d].Interval が異なる: got %d, want %d", i, tc.Interval, want.Interval)
		}
	}
}

// TestLoadConfig_InvalidYAML は不正な YAML に対してエラーが返ることを確認する
func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := os.WriteFile(path, []byte("timers: [invalid: yaml: :::"), 0644); err != nil {
		t.Fatalf("テストファイルの作成に失敗: %v", err)
	}

	_, err := LoadConfig(path)
	if err == nil {
		t.Fatal("不正なYAMLでエラーが発生しなかった")
	}
}

// TestSaveConfig_CreatesDirectory は SaveConfig がディレクトリを自動作成することを確認する
func TestSaveConfig_CreatesDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "deep", "config.yaml")
	cfg := defaultConfig()

	if err := SaveConfig(path, cfg); err != nil {
		t.Fatalf("SaveConfig でディレクトリ作成に失敗: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("ファイルが作成されていない: %v", err)
	}
}

// TestTimersFromConfig は TimerConfig から Timer への変換を確認する
func TestTimersFromConfig(t *testing.T) {
	cfg := &Config{
		Timers: []TimerConfig{
			{Name: "水分補給", Emoji: "💧", Interval: 30},
		},
	}
	timers := timersFromConfig(cfg)
	if len(timers) != 1 {
		t.Fatalf("タイマー数が異なる: got %d, want 1", len(timers))
	}
	if timers[0].Name != "水分補給" {
		t.Errorf("Name が異なる: got %q, want %q", timers[0].Name, "水分補給")
	}
	if timers[0].Interval.Minutes() != 30 {
		t.Errorf("Interval が異なる: got %v, want 30m", timers[0].Interval)
	}
	if timers[0].Remaining != timers[0].Interval {
		t.Errorf("Remaining が Interval と一致しない")
	}
}

// TestTimersFromConfig_NoType は type 省略の既存 YAML が interval 扱いになることを確認する（後方互換）
func TestTimersFromConfig_NoType(t *testing.T) {
	cfg := &Config{
		Timers: []TimerConfig{
			{Name: "水分補給", Emoji: "💧", Interval: 30}, // type 省略
		},
	}
	timers := timersFromConfig(cfg)
	if len(timers) != 1 {
		t.Fatalf("タイマー数が異なる: got %d, want 1", len(timers))
	}
	if timers[0].Type != TimerTypeInterval {
		t.Errorf("type 省略が interval 扱いになっていない: got %q", timers[0].Type)
	}
	if timers[0].IsOnce() {
		t.Error("IsOnce() が true になっている")
	}
}

// TestTimersFromConfig_Once は once タイマーが FireAt を持ち Fired 状態を復元することを確認する
func TestTimersFromConfig_Once(t *testing.T) {
	cfg := &Config{
		Timers: []TimerConfig{
			{Name: "PC終了", Emoji: "💻", Type: TimerTypeOnce, At: "23:00", Fired: true},
		},
	}
	timers := timersFromConfig(cfg)
	if len(timers) != 1 {
		t.Fatalf("タイマー数が異なる: got %d, want 1", len(timers))
	}
	tm := timers[0]
	if !tm.IsOnce() {
		t.Fatal("once タイマーとして認識されていない")
	}
	if tm.FireAt.Hour() != 23 || tm.FireAt.Minute() != 0 {
		t.Errorf("FireAt が異なる: got %02d:%02d, want 23:00", tm.FireAt.Hour(), tm.FireAt.Minute())
	}
	if !tm.Fired || !tm.Notifying {
		t.Errorf("Fired 状態が復元されていない: Fired=%v Notifying=%v", tm.Fired, tm.Notifying)
	}
}

// TestOnceRoundTrip は once タイマーの状態（済み/未）が保存・復元されることを確認する
func TestOnceRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")

	original := []Timer{
		{Name: "PC終了", Emoji: "💻", Type: TimerTypeOnce, FireAt: fireAtToday(23, 0, time.Now()), Fired: true, Notifying: true},
	}
	if err := SaveConfig(path, timersToConfig(original)); err != nil {
		t.Fatalf("SaveConfig に失敗: %v", err)
	}

	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig に失敗: %v", err)
	}
	timers := timersFromConfig(loaded)
	if len(timers) != 1 {
		t.Fatalf("タイマー数が異なる: got %d, want 1", len(timers))
	}
	tm := timers[0]
	if !tm.IsOnce() || tm.FireAt.Hour() != 23 || tm.FireAt.Minute() != 0 {
		t.Errorf("once の発火時刻が復元されていない: %+v", tm)
	}
	if !tm.Fired {
		t.Error("Fired 状態が復元されていない")
	}
}

// TestTimersToConfig は Timer から TimerConfig への変換を確認する
func TestTimersToConfig(t *testing.T) {
	timers := timersFromConfig(defaultConfig())
	cfg := timersToConfig(timers)

	if len(cfg.Timers) != len(timers) {
		t.Fatalf("タイマー数が異なる: got %d, want %d", len(cfg.Timers), len(timers))
	}
	for i, tc := range cfg.Timers {
		want := timers[i]
		if tc.Name != want.Name {
			t.Errorf("タイマー[%d].Name が異なる: got %q, want %q", i, tc.Name, want.Name)
		}
		if float64(tc.Interval) != want.Interval.Minutes() {
			t.Errorf("タイマー[%d].Interval が異なる: got %d, want %v", i, tc.Interval, want.Interval)
		}
	}
}
