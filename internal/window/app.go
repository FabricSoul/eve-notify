// /eve-notify/internal/window/app.go
package window

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/FabricSoul/eve-notify/pkg/character"
	"github.com/FabricSoul/eve-notify/pkg/config"
	"github.com/FabricSoul/eve-notify/pkg/logger"
	"github.com/FabricSoul/eve-notify/pkg/notification"
	"github.com/FabricSoul/eve-notify/pkg/subscription"
)

// ... NewApp and NewMainWindow remain the same ...
func NewApp() fyne.App {
	return app.NewWithID("eve-notify")
}

type refreshResult struct {
	characters []*character.Character
	err        error
}


func NewMainWindow(app fyne.App, charSvc *character.Service, subSvc *subscription.Service, notifSvc *notification.Service) fyne.Window {
	window := app.NewWindow("EVE Notify - Dashboard")

	charData := binding.NewUntypedList()
	rightPane := container.NewVBox(widget.NewLabel("Select a character to configure notifications."))

	// --- PRE-DECLARE VARIABLES for use in closures ---
	var buildRightPane func(char *character.Character)
	var charList *widget.List // <-- THE FIX: Pre-declare charList here

	// --- CONCURRENCY-SAFE REFRESH MECHANISM ---
	resultChan := make(chan refreshResult)

	refreshCharsWorker := func() {
		logger.Sugar.Infoln("Worker: Refreshing character list...")
		chars, err := charSvc.GetCharacters()
		resultChan <- refreshResult{characters: chars, err: err}
	}

	go func() {
		for res := range resultChan {
			if res.err != nil {
				logger.Sugar.Errorf("Failed to refresh characters: %v", res.err)
				dialog.ShowError(res.err, window)
				continue
			}

			logger.Sugar.Infoln("UI: Received new character data, updating list.")

			// charList is now available here because it was pre-declared.
			charList.UnselectAll()

			charItems := make([]interface{}, len(res.characters))
			for i, v := range res.characters {
				charItems[i] = v
			}
			charData.Set(charItems)
		}
	}()
	// --- END OF CONCURRENCY MECHANISM ---

	// Create the list widget and ASSIGN it to our pre-declared variable.
	charList = widget.NewListWithData(charData,
		func() fyne.CanvasObject {
			icon := widget.NewIcon(theme.ConfirmIcon())
			icon.Hide()
			return container.NewHBox(icon, widget.NewLabel("Template Character"))
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			item, _ := i.(binding.Untyped).Get()
			char := item.(*character.Character)
			hbox := o.(*fyne.Container)
			icon := hbox.Objects[0].(*widget.Icon)
			label := hbox.Objects[1].(*widget.Label)
			label.SetText(char.Name)
			if subSvc.IsSubscribed(char.ID) {
				icon.Show()
			} else {
				icon.Hide()
			}
		},
	)

	// ASSIGN the function body to our pre-declared variable.
	buildRightPane = func(char *character.Character) {
		settings, isSubscribed := subSvc.GetSettings(char.ID)
		if !isSubscribed {
			settings = &subscription.NotificationSettings{}
		}
		charNameLabel := widget.NewLabel(fmt.Sprintf("Notifications for: %s", char.Name))
		charNameLabel.TextStyle.Bold = true
		check1 := widget.NewCheck("Alliance chat mentions", func(b bool) { settings.AllianceChat = b })
		check2 := widget.NewCheck("Corp chat mentions", func(b bool) { settings.CorpChat = b })
		check3 := widget.NewCheck("Local chat mentions", func(b bool) { settings.LocalChat = b })
		check4 := widget.NewCheck("Mining storage full", func(b bool) { settings.MiningStorageFull = b })
		check5 := widget.NewCheck("NPC agression stopped", func(b bool) { settings.NpcAggression = b })
		check6 := widget.NewCheck("Player agression", func(b bool) { settings.PlayerAggression = b })
		check7 := widget.NewCheck("Manual Autopilot", func(b bool) { settings.ManualAutopilot = b })
		check1.SetChecked(settings.AllianceChat); check2.SetChecked(settings.CorpChat); check3.SetChecked(settings.LocalChat)
		check4.SetChecked(settings.MiningStorageFull); check5.SetChecked(settings.NpcAggression); check6.SetChecked(settings.PlayerAggression)
		check7.SetChecked(settings.ManualAutopilot)
		formContainer := container.NewVBox(check1, check2, check3, check4, check5, check6, check7)
		setContainerEnabled(formContainer, !isSubscribed)

		var actionButton *widget.Button
		if isSubscribed {
			actionButton = widget.NewButtonWithIcon("Unsubscribe", theme.CancelIcon(), func() {
				subSvc.Unsubscribe(char.ID)
				go refreshCharsWorker()
			})
		} else {
			actionButton = widget.NewButtonWithIcon("Subscribe", theme.ConfirmIcon(), func() {
				subSvc.Subscribe(char.ID, settings)
				notifSvc.Notify("Subscription Active", fmt.Sprintf("Now monitoring notifications for %s.", char.Name), false)
				go refreshCharsWorker()
			})
		}
		rightPane.Objects = []fyne.CanvasObject{
			charNameLabel, widget.NewSeparator(), formContainer, layout.NewSpacer(), actionButton,
		}
		rightPane.Refresh()
	}

	charList.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= charData.Length() {
			return
		}
		item, _ := charData.GetValue(id)
		buildRightPane(item.(*character.Character))
	}

	charList.OnUnselected = func(widget.ListItemID) {
		rightPane.Objects = []fyne.CanvasObject{widget.NewLabel("Select a character to configure notifications.")}
		rightPane.Refresh()
	}

	refreshButton := widget.NewButton("Refresh", func() {
		go refreshCharsWorker()
	})
	leftPane := container.NewBorder(container.NewVBox(widget.NewLabel("Characters"), widget.NewSeparator()), refreshButton, nil, nil, charList)

	split := container.NewHSplit(leftPane, container.NewPadded(rightPane))
	split.Offset = 0.3

	window.SetContent(split)
	window.Resize(fyne.NewSize(1280, 720))

	go refreshCharsWorker()

	window.SetCloseIntercept(func() {
		logger.Sugar.Infoln("Main window closed by user, hiding to tray.")
		window.Hide()
	})
	return window
}



// NewSettingsWindow has been completely redesigned for a professional look.
func NewSettingsWindow(app fyne.App, cfg *config.Service, notifSvc *notification.Service) fyne.Window {
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

	testSoundButton := widget.NewButton("Test Sound", func() {
		logger.Sugar.Infoln("User clicked 'Test Sound' button.")
		// IMPORTANT: Run in a goroutine to avoid freezing the UI.
		go notifSvc.PlaySound()
	})

	pathWidget := container.NewBorder(nil, nil, nil, changePathButton, logPathValue)

	form := widget.NewForm(
		widget.NewFormItem("EVE Log Path", pathWidget),
		widget.NewFormItem("Audio Output", testSoundButton),
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

func setContainerEnabled(c *fyne.Container, enabled bool) {
	for _, obj := range c.Objects {
		if w, ok := obj.(fyne.Disableable); ok {
			if enabled {
				w.Enable()
			} else {
				w.Disable()
			}
		}
	}
}
