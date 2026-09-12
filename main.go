package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// version はビルド時に ldflags で上書き可能
var version = "dev"

// ---- スタイル定義 ----

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B6B")).
			Padding(1, 2)

	timerNormalStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#4ECDC4")).
				Padding(0, 2)

	timerSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#4ECDC4")).
				Padding(0, 2)

	timerAlertStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF6B6B")).
			Background(lipgloss.Color("#2D0A0A")).
			Padding(0, 2)

	timerAlertSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#8B0000")).
				Padding(0, 2)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(1, 2)

	helpGroupStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#4ECDC4"))

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	welcomeStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700")).
			Padding(1, 2)

	formTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFD700")).
			Padding(1, 2)

	formLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")).
			Padding(0, 2)

	formErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Padding(0, 2)
)

// ---- 画面状態 ----

type viewState int

const (
	stateList viewState = iota
	stateAddForm
)

// ---- フォームフィールド ----

type formField int

const (
	fieldType formField = iota // タイマー種別選択（interval / once）
	fieldName
	fieldEmoji
	fieldSchedule // interval なら「分」、once なら「HH:MM」
	fieldAction
	fieldCount // フィールド数（番兵）
)

// ---- メッセージ型 ----

// tickMsg は 1 秒ごとのティックを表す
type tickMsg time.Time

// resetMsg は通知後の自動リセットを表す
type resetMsg struct{ index int }

// ---- ヘルパー ----

// resetTimer は指定インデックスのタイマーを即リセットする。
// once タイマーはリセット対象外（発火済み状態を維持する）。
func (m *model) resetTimer(index int) {
	if index >= 0 && index < len(m.timers) {
		t := &m.timers[index]
		if t.IsOnce() {
			return
		}
		t.Remaining = t.Interval
		t.Notifying = false
	}
}

// adjustVolume はボリュームを delta 分だけ増減し、0.1 刻みに丸めて 0.0〜2.0 にクランプする
func adjustVolume(v, delta float64) float64 {
	return clampVolume(math.Round((v+delta)*10) / 10)
}

// notifyTimer はタイマー発火時の通知を送る
func (m *model) notifyTimer(t *Timer) {
	notifyMsg := t.Name + " の時間です"
	if t.Action != "" {
		notifyMsg += " → " + t.Action
	}
	notify(t.Name, notifyMsg, m.soundEnabled, m.soundName, m.soundVolume)
}

// ---- モデル ----

type model struct {
	// リスト画面
	timers        []Timer
	selectedIndex int

	// 画面状態
	state viewState

	// help エリアの表示/非表示（? でトグル）
	showHelp bool

	// 一時停止中フラグ（p でトグル、SaveConfig には保存しない）
	paused bool

	// 追加フォーム
	formField  formField
	formType   string // fieldType の選択値（TimerTypeInterval / TimerTypeOnce）
	formInputs [4]textinput.Model
	formError  string

	// 設定ファイルパス（終了時セーブ用）
	configPath string

	// 通知設定
	soundEnabled bool
	soundName    string
	soundVolume  float64
}

func initialModel(configPath string) model {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "設定ファイルの読み込みエラー（デフォルト設定で起動します）: %v\n", err)
		cfg = defaultConfig()
	}

	m := model{
		timers:        timersFromConfig(cfg),
		selectedIndex: 0,
		state:         stateList,
		showHelp:      true,
		configPath:    configPath,
		soundEnabled:  cfg.IsSoundEnabled(),
		soundName:     cfg.GetSoundName(),
		soundVolume:   cfg.GetSoundVolume(),
	}
	m.formInputs = newFormInputs()
	return m
}

// textinput 配列のインデックス（fieldType は toggle なので配列に含めない）
const (
	inputName = iota
	inputEmoji
	inputSchedule
	inputAction
)

// inputIndex は formField を textinput 配列のインデックスに変換する。
// fieldType は textinput を持たないため -1 を返す。
func inputIndex(f formField) int {
	switch f {
	case fieldName:
		return inputName
	case fieldEmoji:
		return inputEmoji
	case fieldSchedule:
		return inputSchedule
	case fieldAction:
		return inputAction
	}
	return -1
}

func newFormInputs() [4]textinput.Model {
	nameInput := textinput.New()
	nameInput.Placeholder = "例: 目のストレッチ"
	nameInput.CharLimit = 20

	emojiInput := textinput.New()
	emojiInput.Placeholder = "例: 👁"
	emojiInput.CharLimit = 4

	scheduleInput := textinput.New()
	scheduleInput.Placeholder = "例: 20（分）"
	scheduleInput.CharLimit = 5

	actionInput := textinput.New()
	actionInput.Placeholder = "例: 水を飲む（省略可）"
	actionInput.CharLimit = 40

	return [4]textinput.Model{nameInput, emojiInput, scheduleInput, actionInput}
}

func (m *model) resetForm() {
	m.formInputs = newFormInputs()
	m.formField = fieldType
	m.formType = TimerTypeInterval
	m.formError = ""
}

// ---- Bubble Tea インターフェース ----

func (m model) Init() tea.Cmd {
	return tick()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateList:
			return m.updateList(msg)
		case stateAddForm:
			return m.updateForm(msg)
		}

	case tickMsg:
		cmds := []tea.Cmd{tick()}
		now := time.Time(msg)

		if m.paused {
			return m, tea.Batch(cmds...)
		}

		for i := range m.timers {
			t := &m.timers[i]

			if t.IsOnce() {
				// once: 指定時刻を過ぎたら1回だけ発火。発火後は Notifying のまま永続。
				if t.Fired || t.FireAt.IsZero() || now.Before(t.FireAt) {
					continue
				}
				t.Fired = true
				t.Notifying = true
				m.notifyTimer(t)
				continue
			}

			if t.Notifying {
				continue
			}
			t.Remaining -= time.Second
			if t.Remaining <= 0 {
				t.Remaining = 0
				t.Notifying = true
				m.notifyTimer(t)
				idx := i
				cmds = append(cmds, tea.Tick(10*time.Second, func(_ time.Time) tea.Msg {
					return resetMsg{index: idx}
				}))
			}
		}
		return m, tea.Batch(cmds...)

	case resetMsg:
		m.resetTimer(msg.index)
		return m, nil
	}

	// フォーム画面中のテキスト入力更新（種別選択ステップは textinput を持たない）
	if m.state == stateAddForm {
		idx := inputIndex(m.formField)
		if idx < 0 {
			return m, nil
		}
		var cmd tea.Cmd
		m.formInputs[idx], cmd = m.formInputs[idx].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		if m.selectedIndex > 0 {
			m.selectedIndex--
		}

	case "down", "j":
		if m.selectedIndex < len(m.timers)-1 {
			m.selectedIndex++
		}

	case "a":
		m.state = stateAddForm
		m.resetForm()
		return m, textinput.Blink

	case "d":
		if len(m.timers) > 0 {
			m.timers = append(m.timers[:m.selectedIndex], m.timers[m.selectedIndex+1:]...)
			if m.selectedIndex >= len(m.timers) && m.selectedIndex > 0 {
				m.selectedIndex--
			}
		}

	case "J":
		if m.selectedIndex < len(m.timers)-1 {
			m.timers[m.selectedIndex], m.timers[m.selectedIndex+1] = m.timers[m.selectedIndex+1], m.timers[m.selectedIndex]
			m.selectedIndex++
		}

	case "K":
		if m.selectedIndex > 0 {
			m.timers[m.selectedIndex], m.timers[m.selectedIndex-1] = m.timers[m.selectedIndex-1], m.timers[m.selectedIndex]
			m.selectedIndex--
		}

	case "r":
		m.resetTimer(m.selectedIndex)

	case "+", "=":
		m.soundVolume = adjustVolume(m.soundVolume, 0.1)

	case "-", "_":
		m.soundVolume = adjustVolume(m.soundVolume, -0.1)

	case "p":
		m.paused = !m.paused
		if !m.paused {
			// 一時停止解除時、FireAt が過去になった once タイマーは即発火させる
			now := time.Now()
			for i := range m.timers {
				t := &m.timers[i]
				if t.IsOnce() && !t.Fired && !t.FireAt.IsZero() && !now.Before(t.FireAt) {
					t.Fired = true
					t.Notifying = true
					m.notifyTimer(t)
				}
			}
		}

	case "?":
		m.showHelp = !m.showHelp
	}

	return m, nil
}

func (m model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 種別選択ステップ: left/right/h/l で interval ⇔ once を切り替え
	if m.formField == fieldType {
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.state = stateList
			return m, nil
		case "left", "right", "h", "l", "tab":
			if m.formType == TimerTypeInterval {
				m.formType = TimerTypeOnce
			} else {
				m.formType = TimerTypeInterval
			}
			m.updateSchedulePlaceholder()
			return m, nil
		case "enter":
			m.formError = ""
			m.formField = fieldName
			m.formInputs[inputName].Focus()
			return m, textinput.Blink
		}
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.state = stateList
		return m, nil

	case "enter":
		idx := inputIndex(m.formField)
		currentValue := strings.TrimSpace(m.formInputs[idx].Value())
		if currentValue == "" && m.formField != fieldAction {
			m.formError = "入力が空です"
			return m, nil
		}

		// スケジュールフィールドのバリデーション（種別により内容が変わる）
		if m.formField == fieldSchedule {
			if m.formType == TimerTypeOnce {
				if _, _, ok := parseHHMM(currentValue); !ok {
					m.formError = "時刻を HH:MM 形式（例: 23:00）で入力してください"
					return m, nil
				}
			} else {
				val, err := strconv.Atoi(currentValue)
				if err != nil || val <= 0 {
					m.formError = "正の整数（分）を入力してください"
					return m, nil
				}
			}
		}

		m.formError = ""

		// 最後のフィールド（アクション）なら追加してリストに戻る
		if m.formField == fieldAction {
			m.timers = append(m.timers, m.buildTimerFromForm())
			m.selectedIndex = len(m.timers) - 1
			m.state = stateList
			return m, nil
		}

		// 次のフィールドへ
		m.formInputs[idx].Blur()
		m.formField++
		m.formInputs[inputIndex(m.formField)].Focus()
		return m, textinput.Blink

	default:
		idx := inputIndex(m.formField)
		var cmd tea.Cmd
		m.formInputs[idx], cmd = m.formInputs[idx].Update(msg)
		// interval のスケジュールフィールドは数字のみ許可（once は HH:MM なので : を許容）
		if m.formField == fieldSchedule && m.formType == TimerTypeInterval {
			v := m.formInputs[idx].Value()
			filtered := ""
			for _, c := range v {
				if c >= '0' && c <= '9' {
					filtered += string(c)
				}
			}
			if filtered != v {
				m.formInputs[idx].SetValue(filtered)
			}
		}
		return m, cmd
	}
}

// updateSchedulePlaceholder は種別に応じてスケジュール入力のプレースホルダを更新する
func (m *model) updateSchedulePlaceholder() {
	if m.formType == TimerTypeOnce {
		m.formInputs[inputSchedule].Placeholder = "例: 23:00（HH:MM）"
	} else {
		m.formInputs[inputSchedule].Placeholder = "例: 20（分）"
	}
}

// buildTimerFromForm はフォーム入力から Timer を組み立てる
func (m model) buildTimerFromForm() Timer {
	name := strings.TrimSpace(m.formInputs[inputName].Value())
	emoji := strings.TrimSpace(m.formInputs[inputEmoji].Value())
	schedule := strings.TrimSpace(m.formInputs[inputSchedule].Value())
	action := strings.TrimSpace(m.formInputs[inputAction].Value())

	if m.formType == TimerTypeOnce {
		hour, min, _ := parseHHMM(schedule)
		return Timer{
			Name:   name,
			Emoji:  emoji,
			Action: action,
			Type:   TimerTypeOnce,
			FireAt: fireAtToday(hour, min, time.Now()),
		}
	}

	minutes, _ := strconv.Atoi(schedule)
	d := time.Duration(minutes) * time.Minute
	return Timer{
		Name:      name,
		Emoji:     emoji,
		Action:    action,
		Type:      TimerTypeInterval,
		Interval:  d,
		Remaining: d,
	}
}

func (m model) View() string {
	switch m.state {
	case stateList:
		return m.viewList()
	case stateAddForm:
		return m.viewForm()
	}
	return ""
}

func (m model) viewList() string {
	titleText := "furin - インターバルリマインダー"
	if m.paused {
		titleText += " [一時停止中]"
	}
	title := titleStyle.Render(titleText)

	timerLines := ""
	for i, t := range m.timers {
		selected := i == m.selectedIndex
		var line string
		if t.Notifying {
			if t.Action != "" {
				line = fmt.Sprintf("%s %-12s  時間です！→ %s", t.Emoji, t.Name, t.Action)
			} else {
				line = fmt.Sprintf("%s %-12s  時間です！", t.Emoji, t.Name)
			}
			if selected {
				timerLines += timerAlertSelectedStyle.Render(line) + "\n"
			} else {
				timerLines += timerAlertStyle.Render(line) + "\n"
			}
		} else {
			var status string
			if t.IsOnce() {
				status = formatHHMM(t.FireAt.Hour(), t.FireAt.Minute()) + " 予定"
			} else {
				status = t.Format()
			}
			line = fmt.Sprintf("%s %-12s  %s", t.Emoji, t.Name, status)
			if selected {
				timerLines += timerSelectedStyle.Render(line) + "\n"
			} else {
				timerLines += timerNormalStyle.Render(line) + "\n"
			}
		}
	}

	if len(m.timers) == 0 {
		timerLines = welcomeStyle.Render("まだタイマーがありません。\n'a' キーでタイマーを追加してください。")
	}

	help := m.listHelp()

	return fmt.Sprintf("%s\n%s\n%s\n", title, timerLines, help)
}

// listHelp はリスト画面の help エリアを組み立てる。
// showHelp が false のときは 1 行のヒントのみを返す。
func (m model) listHelp() string {
	if !m.showHelp {
		return helpStyle.Render("?: ヘルプを表示")
	}

	line := func(group, keys string) string {
		return helpGroupStyle.Render(group) + "  " + helpKeyStyle.Render(keys)
	}

	lines := []string{
		line("ナビゲーション", "↑/↓ (k/j): 選択  J/K: 並び替え"),
		line("タイマー操作 ", "a: 追加  d: 削除  r: リセット  p: 一時停止/再開"),
		line("サウンド    ", fmt.Sprintf("+/-: 音量(%.1f)", m.soundVolume)),
		line("アプリ      ", "?: ヘルプ切替  q: 終了"),
	}
	return helpStyle.Render(strings.Join(lines, "\n"))
}

func (m model) scheduleLabel() string {
	if m.formType == TimerTypeOnce {
		return "発火時刻（HH:MM）"
	}
	return "インターバル（分）"
}

func (m model) viewForm() string {
	title := formTitleStyle.Render("タイマーを追加")

	fields := ""
	for f := fieldType; f < fieldCount; f++ {
		label := m.formLabel(f)
		if f == m.formField {
			label = "> " + label
		} else {
			label = "  " + label
		}
		fields += formLabelStyle.Render(label) + "\n"

		if f == fieldType {
			fields += "  " + m.viewTypeToggle() + "\n\n"
			continue
		}
		fields += "  " + m.formInputs[inputIndex(f)].View() + "\n\n"
	}

	errMsg := ""
	if m.formError != "" {
		errMsg = formErrorStyle.Render("✗ "+m.formError) + "\n"
	}

	var help string
	if m.formField == fieldType {
		help = helpStyle.Render("←/→: 種別切替  Enter: 次へ  Esc: キャンセル  Ctrl+C: 終了")
	} else if m.formField == fieldAction {
		help = helpStyle.Render("Enter: 追加を確定  Esc: キャンセル  Ctrl+C: 終了")
	} else {
		help = helpStyle.Render("Enter: 次へ  Esc: キャンセル  Ctrl+C: 終了")
	}

	return fmt.Sprintf("%s\n%s%s%s\n", title, fields, errMsg, help)
}

func (m model) formLabel(f formField) string {
	switch f {
	case fieldType:
		return "種別"
	case fieldName:
		return "名前"
	case fieldEmoji:
		return "絵文字"
	case fieldSchedule:
		return m.scheduleLabel()
	case fieldAction:
		return "完了アクション（省略可）"
	}
	return ""
}

func (m model) viewTypeToggle() string {
	interval := "インターバル"
	once := "ワンショット"
	if m.formType == TimerTypeOnce {
		once = "[ " + once + " ]"
		interval = "  " + interval + "  "
	} else {
		interval = "[ " + interval + " ]"
		once = "  " + once + "  "
	}
	return interval + "   " + once
}

// ---- ティックコマンド ----

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// ---- エントリポイント ----

func main() {
	configFlag := flag.String("config", "", "設定ファイルのパス（デフォルト: ~/.config/furin/config.yaml）")
	showVersion := flag.Bool("version", false, "バージョンを表示して終了")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "furin - ターミナルインターバルリマインダー\n\n")
		fmt.Fprintf(os.Stderr, "使い方: furin [オプション]\n\n")
		fmt.Fprintf(os.Stderr, "オプション:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Printf("furin version %s\n", version)
		return
	}

	configPath := *configFlag
	if configPath == "" {
		var err error
		configPath, err = ConfigPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "設定パスの取得に失敗しました: %v\n", err)
			os.Exit(1)
		}
	}

	p := tea.NewProgram(initialModel(configPath))
	finalModel, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}

	// 終了時に設定を保存
	if m, ok := finalModel.(model); ok {
		cfg := timersToConfig(m.timers)
		cfg.SoundEnabled = boolPtr(m.soundEnabled)
		cfg.SoundName = m.soundName
		cfg.SoundVolume = float64Ptr(m.soundVolume)
		if saveErr := SaveConfig(m.configPath, cfg); saveErr != nil {
			fmt.Fprintf(os.Stderr, "設定の保存に失敗しました: %v\n", saveErr)
		}
	}
}
