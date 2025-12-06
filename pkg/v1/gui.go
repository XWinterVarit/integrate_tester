package v1

import (
	"fmt"
	"image/color"
	"log"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// RunGUI starts the local desktop GUI.
func RunGUI(t *Tester) {
	myApp := app.New()
	myWindow := myApp.NewWindow("Integration Tests")

	// 1. Stage List (Left Pane)
	var stageControls []fyne.CanvasObject
	stageControls = append(stageControls, widget.NewLabelWithStyle("Stages", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))

	for _, stage := range t.Stages {
		stageName := stage.Name // Capture variable

		statusData := binding.NewString()
		_ = statusData.Set("Not Run")

		statusText := canvas.NewText("Not Run", theme.ForegroundColor())
		statusText.TextStyle = fyne.TextStyle{Italic: true}
		statusText.Alignment = fyne.TextAlignTrailing

		statusData.AddListener(binding.NewDataListener(func() {
			val, _ := statusData.Get()
			statusText.Text = val
			if val == "PASSED" {
				statusText.Color = color.NRGBA{R: 0, G: 180, B: 0, A: 255} // Green
			} else if strings.HasPrefix(val, "FAILED") {
				statusText.Color = color.NRGBA{R: 200, G: 0, B: 0, A: 255} // Red
			} else {
				statusText.Color = theme.ForegroundColor()
			}
			statusText.Refresh()
		}))

		runBtn := widget.NewButton("Run", func() {
			_ = statusData.Set("Running...")
			// Run in a separate goroutine to avoid blocking the UI
			go func() {
				err := t.RunStageByName(stageName)

				// Update status on UI thread to ensure canvas.Text modifications are safe
				fyne.Do(func() {
					if err != nil {
						_ = statusData.Set(fmt.Sprintf("FAILED: %v", err))
					} else {
						_ = statusData.Set("PASSED")
					}
				})
			}()
		})

		// Layout for each row: Name [Spacer] Status [Button]
		row := container.NewHBox(
			widget.NewLabel(stageName),
			layout.NewSpacer(),
			statusText,
			runBtn,
		)
		stageControls = append(stageControls, row)
	}

	stageScroll := container.NewScroll(container.NewVBox(stageControls...))

	// 2. Log Pane (Right Pane)
	logsData := binding.NewUntypedList()
	detailData := binding.NewString()
	_ = detailData.Set("Select a log to view details...")

	// Detail View (Bottom)
	detailLabel := widget.NewLabelWithData(detailData)
	detailLabel.Wrapping = fyne.TextWrapWord
	detailScroll := container.NewScroll(detailLabel)

	// Log List (Top)
	logList := widget.NewListWithData(
		logsData,
		func() fyne.CanvasObject {
			return widget.NewLabel("Log Template")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			val, err := i.(binding.Untyped).Get()
			if err != nil {
				return
			}
			entry := val.(LogEntry)

			icon := "🔹"
			switch entry.Type {
			case LogTypeStage:
				icon = "🎬"
			case LogTypeDB:
				icon = "🗄️"
			case LogTypeRequest:
				icon = "🌐"
			case LogTypeMock:
				icon = "🎭"
			case LogTypeApp:
				icon = "🚀"
			case LogTypeExpect:
				icon = "✅"
			case LogTypeInfo:
				icon = "ℹ️"
			}

			label := o.(*widget.Label)
			text := fmt.Sprintf("%s [%s] %s", icon, entry.Type, entry.Summary)

			if entry.Type == LogTypeStage {
				label.TextStyle = fyne.TextStyle{Bold: true}
			} else {
				label.TextStyle = fyne.TextStyle{}
				text = "   " + text // Add left padding for non-stage items
			}
			label.SetText(text)
		},
	)

	logList.OnSelected = func(id widget.ListItemID) {
		val, err := logsData.GetValue(id)
		if err != nil {
			return
		}
		entry := val.(LogEntry)
		_ = detailData.Set(fmt.Sprintf("Type: %s\nSummary: %s\n\n%s", entry.Type, entry.Summary, entry.Detail))
	}

	// Subscribe to logs
	RegisterLogHandler(func(entry LogEntry) {
		_ = logsData.Append(entry)
	})

	// Right Pane Split (List vs Detail)
	rightSplit := container.NewVSplit(logList, detailScroll)
	rightSplit.SetOffset(0.7) // 70% List, 30% Detail

	// Main Split Container
	split := container.NewHSplit(stageScroll, rightSplit)
	split.SetOffset(0.3) // 30% for stages

	myWindow.SetContent(split)
	myWindow.Resize(fyne.NewSize(900, 600))

	log.Println("Starting GUI window...")
	myWindow.ShowAndRun()
}
