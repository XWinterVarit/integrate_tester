package v1

import (
	"fmt"
	"image/color"
	"log"
	"strconv"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// RunGUI starts the local desktop GUI.
func RunGUI(t *Tester) {
	myApp := app.New()
	myWindow := myApp.NewWindow("Integration Tests")

	// --- Shared State ---
	var (
		logs   []LogEntry
		logsMu sync.Mutex

		// Map StageName -> Status String
		stageStatus = make(map[string]string)
		statusMu    sync.Mutex
	)

	// Initialize status
	for _, s := range t.Stages {
		stageStatus[s.Name] = "Not Run"
	}

	// --- Left Pane: Stage & Action Tree ---
	var leftTree *widget.Tree
	leftTree = widget.NewTree(
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			if uid == "" {
				// Root: Return Stage Names
				names := make([]string, len(t.Stages))
				for i, s := range t.Stages {
					names[i] = s.Name
				}
				return names
			}
			// Check if uid is a Stage Name
			// If so, return Action IDs: "StageName:Index"
			actions := GetStageActions(uid)
			if len(actions) > 0 {
				ids := make([]string, len(actions))
				for i := range actions {
					ids[i] = fmt.Sprintf("%s:%d", uid, i)
				}
				return ids
			}
			return nil
		},
		func(uid widget.TreeNodeID) bool {
			// Stages are branches, Actions are leaves
			if uid == "" {
				return true
			}
			if strings.Contains(uid, ":") {
				return false
			} // Action
			return true // Stage
		},
		func(branch bool) fyne.CanvasObject {
			// Template for Node
			return container.NewHBox(
				widget.NewLabel("Template"), // Name/Summary
				layout.NewSpacer(),
				canvas.NewText("Status", theme.ForegroundColor()), // Status (Stage only)
				widget.NewButton("Run", func() {}),                // Run Button
			)
		},
		func(uid widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			box := o.(*fyne.Container)
			label := box.Objects[0].(*widget.Label)
			statusText := box.Objects[2].(*canvas.Text)
			btn := box.Objects[3].(*widget.Button)

			// Determine if Stage or Action
			if strings.Contains(uid, ":") {
				// Action: "StageName:Index"
				parts := strings.SplitN(uid, ":", 2)
				stageName := parts[0]
				idx, _ := strconv.Atoi(parts[1])

				actions := GetStageActions(stageName)
				if idx < len(actions) {
					action := actions[idx]
					label.SetText("  " + action.Summary) // Indent
					label.TextStyle = fyne.TextStyle{Italic: true}
					statusText.Text = "" // No status for actions
					statusText.Refresh()

					btn.SetText("Run")
					btn.OnTapped = func() {
						go func() {
							// Run individual action
							// We don't update stage status for individual run
							Log(LogTypeInfo, "Manual Run: "+action.Summary, "")
							action.Func()
						}()
					}
					btn.Show()
				}
			} else {
				// Stage
				stageName := uid
				label.SetText(stageName)
				label.TextStyle = fyne.TextStyle{Bold: true}

				statusMu.Lock()
				st := stageStatus[stageName]
				statusMu.Unlock()

				statusText.Text = st
				if st == "PASSED" {
					statusText.Color = color.NRGBA{R: 0, G: 180, B: 0, A: 255}
				} else if strings.HasPrefix(st, "FAILED") {
					statusText.Color = color.NRGBA{R: 200, G: 0, B: 0, A: 255}
				} else {
					statusText.Color = theme.ForegroundColor()
				}
				statusText.Refresh()

				btn.SetText("Run Stage")
				btn.OnTapped = func() {
					statusMu.Lock()
					stageStatus[stageName] = "Running..."
					statusMu.Unlock()
					leftTree.RefreshItem(uid) // Refresh to show "Running..."

					go func() {
						err := t.RunStageByName(stageName)
						statusMu.Lock()
						if err != nil {
							stageStatus[stageName] = "FAILED"
						} else {
							stageStatus[stageName] = "PASSED"
						}
						statusMu.Unlock()
						// Refresh GUI
						// Use fyne.Do or just refresh safely
						// RefreshItem is thread-safe? Documentation says "must be called from main thread" usually.
						// The previous code used fyne.Do.
						fyne.Do(func() { leftTree.RefreshItem(uid) })
					}()
				}
				btn.Show()
			}
		},
	)

	// --- Right Pane: Log Tree ---
	// Structure: Root -> Stage Logs -> Operation Logs -> Details
	var rightTree *widget.Tree
	rightTree = widget.NewTree(
		func(uid widget.TreeNodeID) []widget.TreeNodeID {
			logsMu.Lock()
			defer logsMu.Unlock()

			if uid == "" {
				// Find all Stage Logs
				var ids []string
				for i, l := range logs {
					if l.Type == LogTypeStage && strings.HasPrefix(l.Summary, "Running Stage:") {
						ids = append(ids, fmt.Sprintf("%d", i))
					}
				}
				return ids
			}

			// If uid is a number, it's a log index
			if strings.HasSuffix(uid, ":detail") {
				return nil
			}

			idx, err := strconv.Atoi(uid)
			if err != nil {
				return nil
			}

			if idx >= len(logs) {
				return nil
			}
			parentLog := logs[idx]

			// Level 1: Stage -> Operations
			if parentLog.Type == LogTypeStage && strings.HasPrefix(parentLog.Summary, "Running Stage:") {
				var children []string
				// Scan forward until next Stage log
				for i := idx + 1; i < len(logs); i++ {
					l := logs[i]
					if l.Type == LogTypeStage {
						// Stop at next stage event (Start, Passed, Failed)
						// Actually, "PASSED/FAILED" are also Stage logs.
						// We should group them under the "Running" log?
						// Or just list them?
						// Let's stop at next "Running Stage" or end.
						if strings.HasPrefix(l.Summary, "Running Stage:") {
							break
						}
					}
					children = append(children, fmt.Sprintf("%d", i))
				}
				return children
			}

			// Level 2: Operation -> Detail
			if parentLog.Detail != "" {
				// Use a suffix for detail node
				return []string{fmt.Sprintf("%d:detail", idx)}
			}

			return nil
		},
		func(uid widget.TreeNodeID) bool {
			// Check if branch
			if strings.HasSuffix(uid, ":detail") {
				return false
			}

			logsMu.Lock()
			defer logsMu.Unlock()

			if uid == "" {
				return true
			}

			idx, err := strconv.Atoi(uid)
			if err != nil || idx >= len(logs) {
				return false
			}

			l := logs[idx]
			// Stage Logs are branches
			if l.Type == LogTypeStage && strings.HasPrefix(l.Summary, "Running Stage:") {
				return true
			}
			// Logs with details are branches
			if l.Detail != "" {
				return true
			}

			return false
		},
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("Log Entry")
		},
		func(uid widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			label := o.(*widget.Label)

			if strings.HasSuffix(uid, ":detail") {
				// Detail View
				idxStr := strings.TrimSuffix(uid, ":detail")
				idx, _ := strconv.Atoi(idxStr)
				logsMu.Lock()
				detail := logs[idx].Detail
				logsMu.Unlock()
				label.SetText(detail)
				label.TextStyle = fyne.TextStyle{Monospace: true}
				label.Wrapping = fyne.TextWrapWord
				return
			}

			idx, _ := strconv.Atoi(uid)
			logsMu.Lock()
			if idx >= len(logs) {
				logsMu.Unlock()
				return
			}
			entry := logs[idx]
			logsMu.Unlock()

			// Icons
			icon := "🔹"
			switch entry.Type {
			case LogTypeStage:
				icon = "📂"
			case LogTypeDB:
				icon = "🛢️"
			case LogTypeRequest:
				icon = "🌍"
			case LogTypeMock:
				icon = "🤖"
			case LogTypeApp:
				icon = "⚙️"
			case LogTypeExpect:
				icon = "🎯"
			case LogTypeInfo:
				icon = "ℹ️"
			}

			label.SetText(fmt.Sprintf("%s %s", icon, entry.Summary))
			if entry.Type == LogTypeStage {
				label.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				label.TextStyle = fyne.TextStyle{}
			}
			label.Wrapping = fyne.TextWrapOff
		},
	)

	// --- Handlers ---

	// On Action Update (New operation recorded) -> Refresh Left Tree
	RegisterActionUpdateHandler(func() {
		// We need to refresh the node of the current stage
		// Finding the UID of current stage is tricky without passing it.
		// But refreshing Root might be enough?
		// No, RefreshItem("") refreshes structure.
		// We know 'currentStage' in tester, but not here.
		// Let's just refresh root structure.
		fyne.Do(func() {
			if leftTree != nil {
				leftTree.Refresh()
			}
		})
	})

	// On Log Update -> Refresh Right Tree
	RegisterLogHandler(func(entry LogEntry) {
		logsMu.Lock()
		logs = append(logs, entry)
		logsMu.Unlock()
		fyne.Do(func() {
			if rightTree != nil {
				rightTree.Refresh()
			}
		})
		// Auto-scroll? Tree doesn't support easy auto-scroll to bottom.
	})

	// Layout
	split := container.NewHSplit(
		container.NewBorder(widget.NewLabelWithStyle("Test Stages", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}), nil, nil, nil, leftTree),
		container.NewBorder(widget.NewLabelWithStyle("Operation Logs", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}), nil, nil, nil, rightTree),
	)
	split.SetOffset(0.35)

	myWindow.SetContent(split)
	myWindow.Resize(fyne.NewSize(1000, 700))

	log.Println("Starting GUI window...")
	myWindow.ShowAndRun()
}
