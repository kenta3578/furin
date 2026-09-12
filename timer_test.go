package main

import (
	"testing"
	"time"
)

// TestParseHHMM は HH:MM 文字列の解析と検証を確認する
func TestParseHHMM(t *testing.T) {
	cases := []struct {
		in        string
		hour, min int
		ok        bool
	}{
		{"23:00", 23, 0, true},
		{"00:00", 0, 0, true},
		{"09:05", 9, 5, true},
		{"23:59", 23, 59, true},
		{"24:00", 0, 0, false}, // 時が範囲外
		{"12:60", 0, 0, false}, // 分が範囲外
		{"9:05", 0, 0, false},  // ゼロパディングなし（長さ不正）
		{"2300", 0, 0, false},  // コロンなし
		{"ab:cd", 0, 0, false}, // 非数字
		{"", 0, 0, false},      // 空
		{"12:3", 0, 0, false},  // 長さ不正
	}
	for _, c := range cases {
		hour, min, ok := parseHHMM(c.in)
		if ok != c.ok {
			t.Errorf("parseHHMM(%q) ok = %v, want %v", c.in, ok, c.ok)
			continue
		}
		if ok && (hour != c.hour || min != c.min) {
			t.Errorf("parseHHMM(%q) = %d:%d, want %d:%d", c.in, hour, min, c.hour, c.min)
		}
	}
}

// TestFireAtToday は当日の HH:MM を表す time.Time が正しく作られることを確認する
func TestFireAtToday(t *testing.T) {
	now := time.Date(2026, 7, 23, 15, 30, 0, 0, time.Local)
	got := fireAtToday(23, 0, now)
	if got.Hour() != 23 || got.Minute() != 0 {
		t.Errorf("fireAtToday = %02d:%02d, want 23:00", got.Hour(), got.Minute())
	}
	if got.Year() != 2026 || got.Month() != 7 || got.Day() != 23 {
		t.Errorf("fireAtToday date = %v, want 2026-07-23", got)
	}
}
