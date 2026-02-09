package main

import (
	"log"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"percentman/storage"
	"percentman/ui"
)

func main() {
	// Create Fyne app
	a := app.New()

	// Create main window
	window := a.NewWindow("PercentMan - HTTP Client")
	window.Resize(fyne.NewSize(1200, 800))

	// Initialize storage
	store, err := storage.NewStorage()
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Create and build UI
	application := ui.NewApp(a, window, store)
	content := application.BuildUI()

	window.SetContent(content)
	window.ShowAndRun()
}
