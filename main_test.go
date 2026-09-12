package main

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TestOnceFiresAtTime は once タイマーが指定時刻を過ぎたティックで発火し、
// その後のティックで再発火せず Notifying のまま残ることを確認する
func TestOnceFiresAtTime(t *testing.T) {
	fireAt := time.Date(2026, 7, 23, 23, 0, 0, 0, time.Local)
	m := model{
		timers: []Timer{
			{Name: "PC終了", Type: TimerTypeOnce, FireAt: fireAt},
		},
	}

	// 発火前のティック → 発火しない
	before := fireAt.Add(-time.Minute)
	updated, _ := m.Update(tickMsg(before))
	m = updated.(model)
	if m.timers[0].Fired {
		t.Fatal("発火時刻前に発火してしまった")
	}

	// 発火時刻ちょうどのティック → 発火する
	updated, _ = m.Update(tickMsg(fireAt))
	m = updated.(model)
	if !m.timers[0].Fired || !m.timers[0].Notifying {
		t.Fatalf("発火時刻で発火しなかった: Fired=%v Notifying=%v", m.timers[0].Fired, m.timers[0].Notifying)
	}

	// さらに後のティック → 状態は済みのまま維持（再発火しない）
	after := fireAt.Add(time.Hour)
	updated, _ = m.Update(tickMsg(after))
	m = updated.(model)
	if !m.timers[0].Fired || !m.timers[0].Notifying {
		t.Fatal("発火後に済み状態が維持されていない")
	}
}

// TestOnceNotResetByResetTimer は once タイマーが resetTimer で解除されないことを確認する
func TestOnceNotResetByResetTimer(t *testing.T) {
	m := model{
		timers: []Timer{
			{Name: "PC終了", Type: TimerTypeOnce, Fired: true, Notifying: true},
		},
	}
	m.resetTimer(0)
	if !m.timers[0].Notifying {
		t.Error("once タイマーが resetTimer で解除されてしまった")
	}
}

// TestIntervalStillWorks は interval タイマーが従来通り動作することを確認する（後方互換）
func TestIntervalStillWorks(t *testing.T) {
	m := model{
		timers: []Timer{
			{Name: "水分補給", Type: TimerTypeInterval, Interval: 2 * time.Second, Remaining: time.Second},
		},
	}
	updated, _ := m.Update(tickMsg(time.Now()))
	m = updated.(model)
	if !m.timers[0].Notifying {
		t.Fatal("interval タイマーが 0 到達で通知状態にならなかった")
	}
	// reset で復帰する
	m.resetTimer(0)
	if m.timers[0].Notifying || m.timers[0].Remaining != m.timers[0].Interval {
		t.Error("interval タイマーが reset で復帰しなかった")
	}
}

// TestHelpToggle は ? キーで showHelp がトグルし、
// help エリアの表示内容が切り替わることを確認する
func TestHelpToggle(t *testing.T) {
	m := model{state: stateList, showHelp: true}

	full := m.listHelp()
	if !strings.Contains(full, "ナビゲーション") {
		t.Fatalf("showHelp=true でカテゴリ別 help が表示されていない: %q", full)
	}

	updated, _ := m.updateList(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = updated.(model)
	if m.showHelp {
		t.Fatal("? で showHelp が false にトグルされなかった")
	}

	collapsed := m.listHelp()
	if strings.Contains(collapsed, "ナビゲーション") {
		t.Fatalf("showHelp=false でも full help が表示されている: %q", collapsed)
	}
	if !strings.Contains(collapsed, "?") {
		t.Fatalf("折りたたみ時に再表示ヒントが無い: %q", collapsed)
	}
}

// TestWelcomeMessageOnEmpty はタイマー0件のときウェルカムメッセージが表示されることを確認する
func TestWelcomeMessageOnEmpty(t *testing.T) {
	m := model{state: stateList, showHelp: true}
	out := m.viewList()
	if !strings.Contains(out, "'a' キーでタイマーを追加") {
		t.Fatalf("タイマー0件でウェルカムメッセージが表示されていない: %q", out)
	}
}
