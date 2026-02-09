package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	httpclient "percentman/http"
	"percentman/models"
	"percentman/storage"
)

// App represents the main application
type App struct {
	fyneApp    fyne.App
	window     fyne.Window
	storage    *storage.Storage
	httpClient *httpclient.Client

	// Current request state
	currentRequest *models.Request

	// UI Components
	sidebar  *Sidebar
	request  *RequestPanel
	response *ResponsePanel
}

// NewApp creates a new application instance
func NewApp(fyneApp fyne.App, window fyne.Window, store *storage.Storage) *App {
	app := &App{
		fyneApp:        fyneApp,
		window:         window,
		storage:        store,
		httpClient:     httpclient.NewClient(),
		currentRequest: models.NewRequest(),
	}

	// Initialize UI components
	app.sidebar = NewSidebar(app)
	app.request = NewRequestPanel(app)
	app.response = NewResponsePanel(app)

	return app
}

// BuildUI constructs the main UI layout
func (a *App) BuildUI() fyne.CanvasObject {
	// Theme selector (top-right)
	themeSelect := widget.NewSelect([]string{"System", "Dark", "Light"}, func(value string) {
		switch value {
		case "Dark":
			a.fyneApp.Settings().SetTheme(theme.DarkTheme())
		case "Light":
			a.fyneApp.Settings().SetTheme(theme.LightTheme())
		default:
			a.fyneApp.Settings().SetTheme(theme.DefaultTheme())
		}
	})
	themeSelect.SetSelected("System")
	themeSelect.PlaceHolder = "Theme"

	themeLabel := widget.NewLabelWithStyle("Theme:", fyne.TextAlignTrailing, fyne.TextStyle{})
	themeBar := container.NewHBox(
		layout.NewSpacer(),
		themeLabel,
		themeSelect,
	)

	// Left sidebar (templates + history)
	sidebar := a.sidebar.Build()

	// Right side: Request panel (top) + Response panel (bottom)
	requestPanel := a.request.Build()
	responsePanel := a.response.Build()

	// Split request and response vertically (50:50)
	rightSide := container.NewVSplit(requestPanel, responsePanel)
	rightSide.SetOffset(0.5)

	// Right side with theme bar on top
	rightWithTheme := container.NewBorder(themeBar, nil, nil, nil, rightSide)

	// Main layout: sidebar (left) + main content (right)
	mainSplit := container.NewHSplit(sidebar, rightWithTheme)
	mainSplit.SetOffset(0.25) // 25% for sidebar

	return mainSplit
}

// SendRequest executes the current HTTP request
func (a *App) SendRequest() {
	// Update request from UI
	a.request.UpdateRequest(a.currentRequest)

	// Send request
	resp := a.httpClient.SendRequest(a.currentRequest)

	// Display response
	a.response.DisplayResponse(resp)

	// Save to history (only if no error)
	if resp.Error == "" {
		a.storage.AddHistory(a.currentRequest, resp)
		a.sidebar.RefreshHistory()
	}
}

// LoadRequest loads a request into the UI
func (a *App) LoadRequest(req *models.Request) {
	a.currentRequest = req.Clone()
	a.request.LoadRequest(a.currentRequest)
	a.response.Clear()
}

// SaveTemplate saves the current request as a template
func (a *App) SaveTemplate(name string) error {
	a.request.UpdateRequest(a.currentRequest)
	_, err := a.storage.SaveTemplate(name, a.currentRequest)
	if err == nil {
		a.sidebar.RefreshTemplates()
	}
	return err
}

// DeleteTemplate deletes a template
func (a *App) DeleteTemplate(id string) error {
	err := a.storage.DeleteTemplate(id)
	if err == nil {
		a.sidebar.RefreshTemplates()
	}
	return err
}

// ClearHistory clears all history
func (a *App) ClearHistory() error {
	err := a.storage.ClearHistory()
	if err == nil {
		a.sidebar.RefreshHistory()
	}
	return err
}

// GetStorage returns the storage instance
func (a *App) GetStorage() *storage.Storage {
	return a.storage
}

// GetWindow returns the main window
func (a *App) GetWindow() fyne.Window {
	return a.window
}

// ShowSaveTemplateDialog shows a dialog to save the current request as a template
func (a *App) ShowSaveTemplateDialog() {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("Enter template name")

	showSaveDialog(a.window, entry, func(name string) {
		if name != "" {
			a.SaveTemplate(name)
		}
	})
}

func showSaveDialog(window fyne.Window, entry *widget.Entry, onSave func(string)) {
	var popup *widget.PopUp

	titleLabel := widget.NewLabelWithStyle("Save as Template", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	nameLabel := widget.NewLabel("Template Name:")

	// Entry with minimum size for better visibility
	entryContainer := container.NewVBox(entry)

	saveBtn := widget.NewButton("Save", func() {
		if entry.Text != "" {
			onSave(entry.Text)
			popup.Hide()
		}
	})
	saveBtn.Importance = widget.HighImportance

	cancelBtn := widget.NewButton("Cancel", func() {
		popup.Hide()
	})

	buttons := container.NewHBox(
		layout.NewSpacer(),
		cancelBtn,
		saveBtn,
	)

	// Create a padded content area
	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		container.NewVBox(
			nameLabel,
			entryContainer,
		),
		widget.NewSeparator(),
		buttons,
	)

	// Add padding around content
	paddedContent := container.NewPadded(content)

	// Set minimum size for the dialog
	paddedContent.Resize(fyne.NewSize(350, 150))

	popup = widget.NewModalPopUp(paddedContent, window.Canvas())
	popup.Resize(fyne.NewSize(350, 150))
	popup.Show()
}
