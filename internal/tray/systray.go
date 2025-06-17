// /eve-notify/internal/tray/systray.go
package tray

import (
	// "os"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/FabricSoul/eve-notify/pkg/logger"
)

// Setup configures and sets the system tray menu for the application.
func Setup(app fyne.App, mainWindow fyne.Window, settingsWindow fyne.Window) {
	// desk.App is the interface for desktop-specific features.
	// We perform a type assertion to check if the app is running on a desktop.
	if desk, ok := app.(desktop.App); ok {
		logger.Sugar.Infoln("System tray menu supported. Setting up.")

		// Create the menu items.
		// A "Quit" item is automatically added by Fyne.
		menu := fyne.NewMenu("EVE Notify",
			fyne.NewMenuItem("Open", func() {
				mainWindow.Show()
			}),
			fyne.NewMenuItem("Settings",func() {
				settingsWindow.Show()

			}),
		)

		// Set the menu for the system tray.
		desk.SetSystemTrayMenu(menu)
	} else {
		logger.Sugar.Infoln("System tray menu not supported on this platform.")
	}
}
