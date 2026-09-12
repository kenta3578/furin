//go:build darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// notify はタイマー完了時に macOS 通知センターへ通知を送出し、
// サウンドを再生する。非同期で実行されるため TUI をブロックしない。
// エラーが発生しても無視して継続する。
func notify(title, message string, soundEnabled bool, soundName string, volume float64) {
	go func() {
		// macOS 通知センターへ通知
		script := fmt.Sprintf(`display notification "%s" with title "%s"`, message, title)
		cmd := exec.Command("osascript", "-e", script)
		cmd.Run() //nolint:errcheck

		// システムサウンド再生
		if soundEnabled {
			path := soundPath(soundName)
			sound := exec.Command("afplay", "-v", strconv.FormatFloat(volume, 'f', -1, 64), path)
			sound.Run() //nolint:errcheck
		}
	}()
}

// soundPath は指定サウンド名のファイルパスを返す。
// 該当ファイルが存在しない場合はデフォルト Glass にフォールバックする。
func soundPath(name string) string {
	if name == "" {
		name = defaultSoundName
	}
	path := "/System/Library/Sounds/" + name + ".aiff"
	if _, err := os.Stat(path); err != nil {
		return "/System/Library/Sounds/" + defaultSoundName + ".aiff"
	}
	return path
}
