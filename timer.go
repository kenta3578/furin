package main

import "time"

// タイマー種別
const (
	TimerTypeInterval = "interval" // N分おきに繰り返す
	TimerTypeOnce     = "once"     // 指定時刻に1回だけ発火する
	// daily（毎日同時刻）は今回スコープ外
)

// Timer は名前付きタイマーを表す（interval / once）
type Timer struct {
	Name      string
	Emoji     string
	Action    string // 完了時にすべきアクション（任意）
	Type      string // TimerTypeInterval / TimerTypeOnce
	Interval  time.Duration
	Remaining time.Duration
	// FireAt は once タイマーの発火時刻（当日の HH:MM）
	FireAt time.Time
	// Notifying は通知表示中かどうか。once では発火後ずっと true のまま残る
	Notifying bool
	// Fired は once タイマーが発火済みかどうか（「済み」状態の永続化用）
	Fired bool
}

// IsOnce は once タイマーかどうかを返す
func (t Timer) IsOnce() bool {
	return t.Type == TimerTypeOnce
}

// Format は残り時間を MM:SS 形式の文字列に変換する
func (t Timer) Format() string {
	total := int(t.Remaining.Seconds())
	if total < 0 {
		total = 0
	}
	m := total / 60
	s := total % 60
	return formatMMSS(m, s)
}

func formatMMSS(m, s int) string {
	// 手動でゼロパディング（fmt を避けてシンプルに）
	return intToTwoDigit(m) + ":" + intToTwoDigit(s)
}

func intToTwoDigit(n int) string {
	if n < 10 {
		return "0" + intToStr(n)
	}
	return intToStr(n)
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// parseHHMM は "HH:MM" 文字列を時・分に分解して検証する。
// 不正フォーマット（HH:MM 以外、範囲外の値）の場合は ok=false を返す。
func parseHHMM(s string) (hour, min int, ok bool) {
	if len(s) != 5 || s[2] != ':' {
		return 0, 0, false
	}
	for i, c := range s {
		if i == 2 {
			continue
		}
		if c < '0' || c > '9' {
			return 0, 0, false
		}
	}
	hour = int(s[0]-'0')*10 + int(s[1]-'0')
	min = int(s[3]-'0')*10 + int(s[4]-'0')
	if hour > 23 || min > 59 {
		return 0, 0, false
	}
	return hour, min, true
}

// fireAtToday は当日の HH:MM を表す time.Time を返す（ローカルタイム）
func fireAtToday(hour, min int, now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), hour, min, 0, 0, now.Location())
}

// formatHHMM は時・分を "HH:MM" 形式に整形する
func formatHHMM(hour, min int) string {
	return intToTwoDigit(hour) + ":" + intToTwoDigit(min)
}
