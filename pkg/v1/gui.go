package v1

import (
	"fmt"
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// RunGUI starts the local desktop GUI.
func RunGUI(t *Tester) {
	myApp := app.New()
	myWindow := myApp.NewWindow("Integration Tests")

	var stageControls []fyne.CanvasObject

	// Header
	stageControls = append(stageControls, widget.NewLabelWithStyle("Integration Test Stages", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}))

	for _, stage := range t.Stages {
		stageName := stage.Name // Capture variable

		statusLabel := widget.NewLabel("Not Run")
		statusLabel.TextStyle = fyne.TextStyle{Italic: true}

		runBtn := widget.NewButton("Run", func() {
			statusLabel.SetText("Running...")
			// Run in a separate goroutine to avoid blocking the UI
			go func() {
				err := t.RunStageByName(stageName)
				if err != nil {
					statusLabel.SetText(fmt.Sprintf("FAILED: %v", err))
				} else {
					statusLabel.SetText("PASSED")
				}
			}()
		})

		// Layout for each row: Name [Spacer] Status [Button]
		row := container.NewHBox(
			widget.NewLabel(stageName),
			layout.NewSpacer(),
			statusLabel,
			runBtn,
		)
		stageControls = append(stageControls, row)
	}

	// Create a vertical box with all stage rows
	content := container.NewVBox(stageControls...)

	// Wrap in a scroll container in case there are many stages
	scroll := container.NewScroll(content)

	myWindow.SetContent(scroll)
	myWindow.Resize(fyne.NewSize(600, 400))

	log.Println("Starting GUI window...")
	myWindow.ShowAndRun()
}
