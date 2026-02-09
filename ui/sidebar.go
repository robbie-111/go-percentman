package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"percentman/models"
)

// Sidebar represents the left panel with templates and history
type Sidebar struct {
	app *App

	templatesContainer *fyne.Container
	historyContainer   *fyne.Container
}

// NewSidebar creates a new sidebar
func NewSidebar(app *App) *Sidebar {
	return &Sidebar{
		app: app,
	}
}

// Build creates the sidebar UI
func (s *Sidebar) Build() fyne.CanvasObject {
	// Templates section
	templatesTitle := container.NewHBox(
		widget.NewIcon(theme.FolderIcon()),
		widget.NewLabelWithStyle("Templates", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	saveBtn := widget.NewButtonWithIcon("Save Current", theme.ContentAddIcon(), func() {
		s.app.ShowSaveTemplateDialog()
	})

	s.templatesContainer = container.NewVBox()
	s.RefreshTemplates()

	templatesScroll := container.NewVScroll(s.templatesContainer)
	templatesScroll.SetMinSize(fyne.NewSize(200, 150))

	templatesSection := container.NewBorder(
		templatesTitle,
		saveBtn,
		nil, nil,
		templatesScroll,
	)

	// History section
	historyTitle := container.NewHBox(
		widget.NewIcon(theme.HistoryIcon()),
		widget.NewLabelWithStyle("History", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	clearBtn := widget.NewButtonWithIcon("Clear All", theme.DeleteIcon(), func() {
		s.app.ClearHistory()
	})

	s.historyContainer = container.NewVBox()
	s.RefreshHistory()

	historyScroll := container.NewVScroll(s.historyContainer)
	historyScroll.SetMinSize(fyne.NewSize(200, 150))

	historySection := container.NewBorder(
		historyTitle,
		clearBtn,
		nil, nil,
		historyScroll,
	)

	// Split templates and history vertically (50:50)
	split := container.NewVSplit(templatesSection, historySection)
	split.SetOffset(0.5)

	return split
}

// RefreshTemplates refreshes the templates list
func (s *Sidebar) RefreshTemplates() {
	s.templatesContainer.RemoveAll()

	templates := s.app.GetStorage().GetTemplates()

	if len(templates) == 0 {
		s.templatesContainer.Add(widget.NewLabel("No templates saved"))
	} else {
		for _, t := range templates {
			template := t // capture for closure
			item := s.createTemplateItem(&template)
			s.templatesContainer.Add(item)
			s.templatesContainer.Add(widget.NewSeparator())
		}
	}

	s.templatesContainer.Refresh()
}

// createTemplateItem creates a template list item (name only, single line)
func (s *Sidebar) createTemplateItem(t *models.Template) fyne.CanvasObject {
	// Single line: Template name (bold) + delete button
	nameLabel := widget.NewLabelWithStyle(t.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	nameLabel.Truncation = fyne.TextTruncateEllipsis

	deleteBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		s.app.DeleteTemplate(t.ID)
	})
	deleteBtn.Importance = widget.LowImportance

	content := container.NewBorder(nil, nil, nil, deleteBtn, nameLabel)

	// Make the whole row clickable with tooltip (full URL)
	tooltipText := fmt.Sprintf("%s %s", t.Request.Method, t.Request.URL)
	clickable := NewClickableContainer(content, func() {
		s.app.LoadRequest(&t.Request)
	}, tooltipText, s.app.GetWindow())

	return clickable
}

// RefreshHistory refreshes the history list
func (s *Sidebar) RefreshHistory() {
	s.historyContainer.RemoveAll()

	history := s.app.GetStorage().GetHistory()

	if len(history) == 0 {
		s.historyContainer.Add(widget.NewLabel("No history yet"))
	} else {
		for _, h := range history {
			item := h // capture for closure
			historyItem := s.createHistoryItem(&item)
			s.historyContainer.Add(historyItem)
			s.historyContainer.Add(widget.NewSeparator())
		}
	}

	s.historyContainer.Refresh()
}

// createHistoryItem creates a history list item with 2-line layout
func (s *Sidebar) createHistoryItem(h *models.HistoryItem) fyne.CanvasObject {
	// Line 1: Method + Full URL
	methodLabel := widget.NewLabelWithStyle(
		h.Request.Method,
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	// Show full URL (no truncation text, just ellipsis if needed)
	urlLabel := widget.NewLabel(h.Request.URL)
	urlLabel.Truncation = fyne.TextTruncateEllipsis

	line1 := container.NewBorder(nil, nil, methodLabel, nil, urlLabel)

	// Line 2: Status code + response time
	statusText := fmt.Sprintf("%d %s", h.Response.StatusCode, getStatusText(h.Response.StatusCode))
	statusLabel := widget.NewLabel(statusText)
	if h.Response.StatusCode >= 200 && h.Response.StatusCode < 300 {
		statusLabel.Importance = widget.SuccessImportance
	} else if h.Response.StatusCode >= 400 {
		statusLabel.Importance = widget.DangerImportance
	} else {
		statusLabel.Importance = widget.WarningImportance
	}

	timeLabel := widget.NewLabel(fmt.Sprintf("%dms", h.Response.ResponseTime.Milliseconds()))
	timeLabel.Importance = widget.LowImportance

	line2 := container.NewHBox(statusLabel, widget.NewLabel("-"), timeLabel)

	// Combined 2-line layout
	content := container.NewVBox(line1, line2)

	// Make clickable with tooltip (full URL)
	tooltipText := fmt.Sprintf("%s %s", h.Request.Method, h.Request.URL)
	clickable := NewClickableContainer(content, func() {
		s.app.LoadRequest(&h.Request)
	}, tooltipText, s.app.GetWindow())

	return clickable
}

// getStatusText returns a short status text for common status codes
func getStatusText(code int) string {
	switch code {
	case 200:
		return "OK"
	case 201:
		return "Created"
	case 204:
		return "No Content"
	case 400:
		return "Bad Request"
	case 401:
		return "Unauthorized"
	case 403:
		return "Forbidden"
	case 404:
		return "Not Found"
	case 500:
		return "Server Error"
	default:
		return ""
	}
}

// ClickableContainer is a container that responds to taps and shows tooltip on hover
type ClickableContainer struct {
	widget.BaseWidget
	content    fyne.CanvasObject
	onTapped   func()
	tooltip    string
	window     fyne.Window
	popup      *widget.PopUp
	hoverTimer *time.Timer
	isHovering bool
}

// NewClickableContainer creates a new clickable container with tooltip support
func NewClickableContainer(content fyne.CanvasObject, onTapped func(), tooltip string, window fyne.Window) *ClickableContainer {
	c := &ClickableContainer{
		content:  content,
		onTapped: onTapped,
		tooltip:  tooltip,
		window:   window,
	}
	c.ExtendBaseWidget(c)
	return c
}

// CreateRenderer returns the renderer for this widget
func (c *ClickableContainer) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(c.content)
}

// Tapped handles tap events
func (c *ClickableContainer) Tapped(*fyne.PointEvent) {
	if c.onTapped != nil {
		c.onTapped()
	}
}

// TappedSecondary handles secondary tap events (right-click)
func (c *ClickableContainer) TappedSecondary(*fyne.PointEvent) {}

// MouseIn starts tooltip timer when mouse enters (implements desktop.Hoverable)
func (c *ClickableContainer) MouseIn(e *desktop.MouseEvent) {
	c.isHovering = true

	if c.tooltip == "" || c.window == nil {
		return
	}

	// Store position for later use
	pos := e.AbsolutePosition

	// Start 2-second timer for tooltip
	c.hoverTimer = time.AfterFunc(2*time.Second, func() {
		if !c.isHovering {
			return
		}

		// Show tooltip on main thread
		c.showTooltip(pos)
	})
}

// showTooltip displays the tooltip popup
func (c *ClickableContainer) showTooltip(pos fyne.Position) {
	if c.popup != nil {
		c.popup.Hide()
	}

	// Create horizontal tooltip label
	tooltipLabel := widget.NewLabel(c.tooltip)

	// Wrap in padded container for better appearance
	tooltipContent := container.NewPadded(tooltipLabel)

	c.popup = widget.NewPopUp(tooltipContent, c.window.Canvas())

	// Position tooltip below and to the right of mouse cursor
	c.popup.ShowAtPosition(fyne.NewPos(pos.X+10, pos.Y+20))
}

// MouseMoved handles mouse movement (implements desktop.Hoverable)
func (c *ClickableContainer) MouseMoved(*desktop.MouseEvent) {}

// MouseOut hides tooltip and cancels timer when mouse leaves (implements desktop.Hoverable)
func (c *ClickableContainer) MouseOut() {
	c.isHovering = false

	// Cancel pending timer
	if c.hoverTimer != nil {
		c.hoverTimer.Stop()
		c.hoverTimer = nil
	}

	// Hide popup if showing
	if c.popup != nil {
		c.popup.Hide()
		c.popup = nil
	}
}

// Verify interface implementation
var _ desktop.Hoverable = (*ClickableContainer)(nil)
