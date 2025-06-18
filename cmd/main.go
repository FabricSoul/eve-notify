// /eve-notify/cmd/main.go
package main

import (
	_ "image/png"

	"github.com/FabricSoul/eve-notify/internal/tray"
	"github.com/FabricSoul/eve-notify/internal/window"
	"github.com/FabricSoul/eve-notify/pkg/character"
	"github.com/FabricSoul/eve-notify/pkg/config"
	"github.com/FabricSoul/eve-notify/pkg/logger"
	"github.com/FabricSoul/eve-notify/pkg/subscription"
	// We no longer need to import "github.com/getlantern/systray"
)

func main() {
	cleanup := logger.Init()
	defer cleanup()

	// 1. Create the Fyne app and window.
	logger.Sugar.Infoln("Starting the Fyne app")
	mainApp := window.NewApp()

	configService := config.NewService(mainApp)
	subService := subscription.NewService()
	characaterService := character.NewService(mainApp,  configService, subService)

	configService.Init()

	mainWindow := window.NewMainWindow(mainApp, characaterService, subService)
	settingsWindow := window.NewSettingsWindow(mainApp, configService)



	// 2. Set up the system tray menu using our refactored tray package.
	tray.Setup(mainApp, mainWindow, settingsWindow)

	// 3. Hide the window initially to start as a tray-only application.
	// You can change this to mainWindow.Show() if you want it visible on startup.
	mainWindow.Hide()

	// 4. Run the Fyne application. This is a blocking call that runs the event
	// loop for BOTH the window and the system tray menu. It will exit when
	// app.Quit() is called (e.g., from the default "Quit" tray menu item).
	mainApp.Run()

	// When mainApp.Run() returns, the app is closing.
	logger.Sugar.Infoln("Fyne app has quit.")
	// No need to call systray.Quit() anymore!
}
