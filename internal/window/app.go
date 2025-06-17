// /eve-notify/internal/window/app.go
package window

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fmt"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/FabricSoul/eve-notify/pkg/character"
	"github.com/FabricSoul/eve-notify/pkg/config"
	"github.com/FabricSoul/eve-notify/pkg/logger"
)

// ... NewApp and NewMainWindow remain the same ...
func NewApp() fyne.App {
	return app.NewWithID("eve-notify")
}

func NewMainWindow(app fyne.App, charSvc *character.Service) fyne.Window {
	window := app.NewWindow("EVE Notify - Dashboard")

	// --- DATA & STATE ---
	subscriptions := make(map[int64]bool)
	charData := binding.NewUntypedList()

	// --- RIGHT PANE (FORM) ---
	rightPane := container.NewVBox(widget.NewLabel("Select a character from the list to configure notifications."))

	// --- LEFT PANE (LIST) ---
	charList := widget.NewListWithData(charData,
		func() fyne.CanvasObject {
			// This creates an object of type *widget.Label
			return widget.NewLabel("Template Character Name")
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			item, _ := i.(binding.Untyped).Get()
			char := item.(*character.Character)
			// --- THE FIX IS HERE ---
			// We assert the type is *widget.Label, not *widget.NewLabel
			o.(*widget.Label).SetText(char.Name)
			// --- END OF FIX ---
		},
	)

	charList.OnSelected = func(id widget.ListItemID) {
		// Prevent crash if selection is cleared
		if id < 0 || id >= charData.Length() {
			rightPane.Objects = []fyne.CanvasObject{widget.NewLabel("Select a character from the list to configure notifications.")}
			rightPane.Refresh()
			return
		}

		item, _ := charData.GetValue(id)
		char := item.(*character.Character)
		logger.Sugar.Infof("User selected character: %s (%d)", char.Name, char.ID)

		charNameLabel := widget.NewLabel(fmt.Sprintf("Notifications for: %s", char.Name))
		charNameLabel.TextStyle.Bold = true

		check1 := widget.NewCheck("Alliance chat mentions", nil)
		check2 := widget.NewCheck("Corp chat mentions", nil)
		check3 := widget.NewCheck("Local chat mentions", nil)
		check4 := widget.NewCheck("Mining storage full", nil)
		check5 := widget.NewCheck("NPC agression stopped", nil)
		check6 := widget.NewCheck("Player agression", nil)
		check7 := widget.NewCheck("Manual Autopilot", nil)


		subscribeButton := widget.NewButton("Subscribe", func() {
			subscriptions[char.ID] = true
			logger.Sugar.Infof("Subscribed to notifications for %s", char.Name)
			dialog.ShowInformation("Subscribed", "You will now receive notifications for "+char.Name, window)
		})
		cancelButton := widget.NewButton("Cancel Subscription", func() {
			delete(subscriptions, char.ID)
			logger.Sugar.Infof("Cancelled subscription for %s", char.Name)
			dialog.ShowInformation("Cancelled", "You will no longer receive notifications for "+char.Name, window)
		})

		rightPane.Objects = []fyne.CanvasObject{
			charNameLabel,
			widget.NewSeparator(),
			check1,
			check2,
			check3,
			check4,
			check5,
			check6,
			check7,
			layout.NewSpacer(),
			container.NewGridWithColumns(2, subscribeButton, cancelButton),
		}
		rightPane.Refresh()
	}
	// When selection is cleared, reset the right pane
	charList.OnUnselected = func(id widget.ListItemID) {
		charList.OnSelected(-1) // Call the selection handler with an invalid ID
	}

	refreshChars := func() {
		logger.Sugar.Infoln("Refreshing character list...")
		// Clear selection before refresh to avoid panics on list change
		charList.UnselectAll()

		chars, err := charSvc.GetCharacters()
		if err != nil {
			logger.Sugar.Errorf("Failed to refresh characters: %v", err)
			dialog.ShowError(err, window)
			return
		}
		charItems := make([]interface{}, len(chars))
		for i, v := range chars {
			charItems[i] = v
		}
		charData.Set(charItems)
	}

	refreshButton := widget.NewButton("Refresh", refreshChars)
	leftPane := container.NewBorder(container.NewVBox(widget.NewLabel("Characters"), widget.NewSeparator()), refreshButton, nil, nil, charList)

	split := container.NewHSplit(leftPane, container.NewPadded(rightPane))
	split.Offset = 0.3

	window.SetContent(split)
	window.Resize(fyne.NewSize(1280, 720))

	go refreshChars()

	window.SetCloseIntercept(func() {
		logger.Sugar.Infoln("Main window closed by user, hiding to tray.")
		window.Hide()
	})
	return window
}

// NewSettingsWindow has been completely redesigned for a professional look.
func NewSettingsWindow(app fyne.App, cfg *config.Service) fyne.Window {
	logger.Sugar.Debugln("Creating settings window UI.")
	window := app.NewWindow("Settings")

	// --- WIDGETS ---
	logPath := cfg.GetLogPath()
	if logPath == "" {
		logPath = "Not Set"
	}
	logPathValue := widget.NewEntry()
	if logPath == "" {
		logPathValue.SetPlaceHolder("Not Set")
	} else {
		logPathValue.SetText(logPath)
	}
	logPathValue.Disable()

	changePathButton := widget.NewButton("Change...", func() {
		logger.Sugar.Infoln("User clicked 'Change' log path button.")
		folderDialog := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				logger.Sugar.Errorf("Error from folder dialog: %v", err)
				return
			}
			if uri == nil {
				logger.Sugar.Infoln("User cancelled folder selection.")
				return
			}
			selectedPath := uri.Path()
			cfg.SetLogPath(selectedPath)
			logPathValue.SetText(selectedPath)
		}, window)
		folderDialog.Resize(fyne.NewSize(960, 540))
		folderDialog.Show()
	})

	pathWidget := container.NewBorder(nil, nil, nil, changePathButton, logPathValue)

	form := widget.NewForm(
		widget.NewFormItem("EVE Log Path", pathWidget),
	)

	btnClose := widget.NewButton("Close", func() {
		logger.Sugar.Infoln("User closed settings window.")
		window.Hide()
	})
	btnDefault := widget.NewButton("Restore Default", func() {
		logger.Sugar.Infoln("User clicked 'Restore Default'.")
		 restoredPath := cfg.RestoreDefaultLogPath()

    if restoredPath == "" {
			logPathValue.SetText("Not Set")
    } else {
			logPathValue.SetText(restoredPath)
    }
	})

	bottomBar := container.NewHBox(layout.NewSpacer(), btnDefault, btnClose)

content := container.NewPadded(
		container.NewBorder(form, bottomBar, nil, nil),
	)
	window.SetContent(content)
	window.Resize(fyne.NewSize(960, 540))
	window.SetFixedSize(true)

	return window
}
