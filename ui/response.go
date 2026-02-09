package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	httpclient "percentman/http"
	"percentman/models"
)

// ResponsePanel represents the response display panel
type ResponsePanel struct {
	app *App

	statusLabel *widget.Label
	timeLabel   *widget.Label
	headersText *widget.Entry
	bodyText    *widget.Entry
	lastHeaders string
	lastBody    string
}

// NewResponsePanel creates a new response panel
func NewResponsePanel(app *App) *ResponsePanel {
	return &ResponsePanel{
		app: app,
	}
}

// Build creates the response panel UI
func (r *ResponsePanel) Build() fyne.CanvasObject {
	// Status and time labels
	r.statusLabel = widget.NewLabel("Status: -")
	r.timeLabel = widget.NewLabel("Time: -")

	statusBar := container.NewHBox(
		widget.NewIcon(theme.InfoIcon()),
		widget.NewLabelWithStyle("Response", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		r.statusLabel,
		widget.NewSeparator(),
		r.timeLabel,
	)

	// Response headers - enabled for better readability
	r.headersText = widget.NewMultiLineEntry()
	r.headersText.SetPlaceHolder("Response headers will appear here")
	r.headersText.Wrapping = fyne.TextWrapWord
	// Make it read-only by reverting changes
	r.headersText.OnChanged = func(s string) {
		if s != r.lastHeaders {
			r.headersText.SetText(r.lastHeaders)
		}
	}

	headersSection := container.NewBorder(
		widget.NewLabel("Headers"),
		nil, nil, nil,
		r.headersText,
	)

	// Response body - enabled for better readability
	r.bodyText = widget.NewMultiLineEntry()
	r.bodyText.SetPlaceHolder("Response body will appear here")
	r.bodyText.Wrapping = fyne.TextWrapWord
	// Make it read-only by reverting changes
	r.bodyText.OnChanged = func(s string) {
		if s != r.lastBody {
			r.bodyText.SetText(r.lastBody)
		}
	}

	bodySection := container.NewBorder(
		widget.NewLabel("Body"),
		nil, nil, nil,
		r.bodyText,
	)

	// Tabs for Headers and Body
	tabs := container.NewAppTabs(
		container.NewTabItem("Body", bodySection),
		container.NewTabItem("Headers", headersSection),
	)

	return container.NewBorder(
		statusBar,
		nil, nil, nil,
		tabs,
	)
}

// DisplayResponse displays the HTTP response
func (r *ResponsePanel) DisplayResponse(resp *models.Response) {
	if resp.Error != "" {
		r.statusLabel.SetText("Error: " + resp.Error)
		r.statusLabel.Importance = widget.DangerImportance
		r.timeLabel.SetText("Time: -")
		r.lastBody = ""
		r.lastHeaders = ""
		r.bodyText.SetText("")
		r.headersText.SetText("")
		return
	}

	// Status
	r.statusLabel.SetText(fmt.Sprintf("Status: %s", resp.Status))
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		r.statusLabel.Importance = widget.SuccessImportance
	} else if resp.StatusCode >= 400 {
		r.statusLabel.Importance = widget.DangerImportance
	} else {
		r.statusLabel.Importance = widget.MediumImportance
	}

	// Time
	r.timeLabel.SetText(fmt.Sprintf("Time: %dms", resp.ResponseTime.Milliseconds()))

	// Headers
	headersStr := ""
	for k, v := range resp.Headers {
		headersStr += fmt.Sprintf("%s: %s\n", k, v)
	}
	r.lastHeaders = headersStr
	r.headersText.SetText(headersStr)

	// Body (format JSON if possible)
	body := resp.Body
	if httpclient.IsJSON(body) {
		body = httpclient.FormatJSON(body)
	}
	r.lastBody = body
	r.bodyText.SetText(body)
}

// Clear clears the response panel
func (r *ResponsePanel) Clear() {
	r.statusLabel.SetText("Status: -")
	r.statusLabel.Importance = widget.MediumImportance
	r.timeLabel.SetText("Time: -")
	r.lastHeaders = ""
	r.lastBody = ""
	r.headersText.SetText("")
	r.bodyText.SetText("")
}
