package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"percentman/models"
)

// RequestPanel represents the request input panel
type RequestPanel struct {
	app *App

	methodSelect     *widget.Select
	urlEntry         *widget.Entry
	headersContainer *fyne.Container
	bodyEntry        *widget.Entry
	headers          []headerRow
}

type headerRow struct {
	keyEntry   *widget.Entry
	valueEntry *widget.Entry
	enabled    *widget.Check
}

// NewRequestPanel creates a new request panel
func NewRequestPanel(app *App) *RequestPanel {
	return &RequestPanel{
		app:     app,
		headers: []headerRow{},
	}
}

// Build creates the request panel UI
func (r *RequestPanel) Build() fyne.CanvasObject {
	// Method selector
	r.methodSelect = widget.NewSelect(
		[]string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		func(value string) {},
	)
	r.methodSelect.SetSelected("GET")

	// URL entry
	r.urlEntry = widget.NewEntry()
	r.urlEntry.SetPlaceHolder("Enter URL (e.g., https://api.example.com/users)")

	// Send button
	sendBtn := widget.NewButtonWithIcon("Send", theme.MediaPlayIcon(), func() {
		r.app.SendRequest()
	})
	sendBtn.Importance = widget.HighImportance

	// Top bar: Method + URL + Send
	urlContainer := container.NewBorder(nil, nil, r.methodSelect, sendBtn, r.urlEntry)

	// Headers section
	headersLabel := widget.NewLabelWithStyle("Headers", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	addHeaderBtn := widget.NewButtonWithIcon("Add Header", theme.ContentAddIcon(), func() {
		r.addHeaderRow("", "", true)
	})

	r.headersContainer = container.NewVBox()
	// Add one default header row
	r.addHeaderRow("Content-Type", "application/json", true)

	headersScroll := container.NewVScroll(r.headersContainer)
	headersScroll.SetMinSize(fyne.NewSize(0, 100))

	headersSection := container.NewBorder(
		container.NewHBox(headersLabel, addHeaderBtn),
		nil, nil, nil,
		headersScroll,
	)

	// Body section
	bodyLabel := widget.NewLabelWithStyle("Body", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	r.bodyEntry = widget.NewMultiLineEntry()
	r.bodyEntry.SetPlaceHolder("Request body (JSON)")
	r.bodyEntry.SetMinRowsVisible(5)

	bodySection := container.NewBorder(bodyLabel, nil, nil, nil, r.bodyEntry)

	// Tabs for Headers and Body
	tabs := container.NewAppTabs(
		container.NewTabItem("Headers", headersSection),
		container.NewTabItem("Body", bodySection),
	)

	// Main layout
	return container.NewBorder(
		urlContainer,
		nil, nil, nil,
		tabs,
	)
}

// addHeaderRow adds a new header row to the headers container
func (r *RequestPanel) addHeaderRow(key, value string, enabled bool) {
	keyEntry := widget.NewEntry()
	keyEntry.SetPlaceHolder("Header name")
	keyEntry.SetText(key)

	valueEntry := widget.NewEntry()
	valueEntry.SetPlaceHolder("Header value")
	valueEntry.SetText(value)

	enabledCheck := widget.NewCheck("", nil)
	enabledCheck.SetChecked(enabled)

	row := headerRow{
		keyEntry:   keyEntry,
		valueEntry: valueEntry,
		enabled:    enabledCheck,
	}
	r.headers = append(r.headers, row)

	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		r.removeHeaderRow(len(r.headers) - 1)
	})
	deleteBtn.Importance = widget.LowImportance

	rowContainer := container.NewBorder(
		nil, nil,
		enabledCheck,
		deleteBtn,
		container.NewGridWithColumns(2, keyEntry, valueEntry),
	)

	r.headersContainer.Add(rowContainer)
	r.headersContainer.Refresh()
}

// removeHeaderRow removes a header row
func (r *RequestPanel) removeHeaderRow(index int) {
	if index < 0 || index >= len(r.headers) {
		return
	}

	// Remove from headers slice
	r.headers = append(r.headers[:index], r.headers[index+1:]...)

	// Rebuild headers container
	r.headersContainer.RemoveAll()
	for i, h := range r.headers {
		idx := i
		deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
			r.removeHeaderRow(idx)
		})
		deleteBtn.Importance = widget.LowImportance

		rowContainer := container.NewBorder(
			nil, nil,
			h.enabled,
			deleteBtn,
			container.NewGridWithColumns(2, h.keyEntry, h.valueEntry),
		)
		r.headersContainer.Add(rowContainer)
	}
	r.headersContainer.Refresh()
}

// UpdateRequest updates the request model from UI state
func (r *RequestPanel) UpdateRequest(req *models.Request) {
	req.Method = r.methodSelect.Selected
	req.URL = r.urlEntry.Text
	req.Body = r.bodyEntry.Text

	req.Headers = []models.Header{}
	for _, h := range r.headers {
		if h.keyEntry.Text != "" {
			req.Headers = append(req.Headers, models.Header{
				Key:     h.keyEntry.Text,
				Value:   h.valueEntry.Text,
				Enabled: h.enabled.Checked,
			})
		}
	}
}

// LoadRequest loads a request into the UI
func (r *RequestPanel) LoadRequest(req *models.Request) {
	r.methodSelect.SetSelected(req.Method)
	r.urlEntry.SetText(req.URL)
	r.bodyEntry.SetText(req.Body)

	// Clear and rebuild headers
	r.headers = []headerRow{}
	r.headersContainer.RemoveAll()

	if len(req.Headers) == 0 {
		r.addHeaderRow("", "", true)
	} else {
		for _, h := range req.Headers {
			r.addHeaderRow(h.Key, h.Value, h.Enabled)
		}
	}
}
