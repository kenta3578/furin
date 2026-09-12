//go:build !darwin && !linux

package main

// notify は macOS / Linux 以外の環境では何もしない（no-op）。
// TUI がクラッシュしないようにスタブとして定義する。
func notify(title, message string, soundEnabled bool, soundName string, volume float64) {
	// 非 macOS 環境ではスキップ
}
