//go:build linux

package main

import (
	"os"
	"os/exec"
)

// notify はタイマー完了時に Linux デスクトップへ通知を送出し、
// サウンドを再生する。非同期で実行されるため TUI をブロックしない。
// 通知は notify-send（libnotify）、サウンドは paplay → aplay の順で
// フォールバックする。該当コマンドやサウンドが無い場合は無音で継続する。
func notify(title, message string, soundEnabled bool, soundName string, volume float64) {
	go func() {
		if path, err := exec.LookPath("notify-send"); err == nil {
			exec.Command(path, title, message).Run() //nolint:errcheck
		}

		if soundEnabled {
			playSound()
		}
	}()
}

// playSound は標準サウンドファイルを paplay → aplay の順で再生する。
// 再生コマンドまたはサウンドファイルが存在しない場合はサイレントで継続する。
func playSound() {
	path := linuxSoundPath()
	if path == "" {
		return
	}

	if p, err := exec.LookPath("paplay"); err == nil {
		exec.Command(p, path).Run() //nolint:errcheck
		return
	}
	if p, err := exec.LookPath("aplay"); err == nil {
		exec.Command(p, path).Run() //nolint:errcheck
	}
}

// linuxSoundPath は freedesktop 標準サウンドのパスを探索する。
// 存在しない場合は空文字を返し、呼び出し側は無音で継続する。
func linuxSoundPath() string {
	candidates := []string{
		"/usr/share/sounds/freedesktop/stereo/complete.oga",
		"/usr/share/sounds/freedesktop/stereo/bell.oga",
		"/usr/share/sounds/alsa/Front_Center.wav",
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}
