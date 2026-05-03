package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// iconTooltipButton is a compact icon button that shows a small hover tooltip.
type iconTooltipButton struct {
	widget.Button
	tooltip string
	popup   *widget.PopUp
}

func newIconTooltipButton(icon fyne.Resource, tooltip string, tapped func()) *iconTooltipButton {
	b := &iconTooltipButton{tooltip: strings.TrimSpace(tooltip)}
	b.Text = ""
	b.Icon = icon
	b.Importance = widget.LowImportance
	b.OnTapped = tapped
	b.ExtendBaseWidget(b)
	return b
}

func (b *iconTooltipButton) MouseIn(ev *desktop.MouseEvent) {
	b.Button.MouseIn(ev)
	b.showTooltip()
}

func (b *iconTooltipButton) MouseMoved(ev *desktop.MouseEvent) {
	b.Button.MouseMoved(ev)
	b.showTooltip()
}

func (b *iconTooltipButton) MouseOut() {
	b.hideTooltip()
	b.Button.MouseOut()
}

func (b *iconTooltipButton) Tapped(ev *fyne.PointEvent) {
	b.hideTooltip()
	b.Button.Tapped(ev)
}

func (b *iconTooltipButton) showTooltip() {
	if b.Disabled() || b.tooltip == "" {
		b.hideTooltip()
		return
	}
	drv := fyne.CurrentApp().Driver()
	canvas := drv.CanvasForObject(b)
	if canvas == nil {
		return
	}
	if b.popup == nil {
		label := widget.NewLabel(b.tooltip)
		label.Wrapping = fyne.TextWrapOff
		b.popup = widget.NewPopUp(container.NewPadded(label), canvas)
	}
	abs := drv.AbsolutePositionForObject(b)
	b.popup.Move(fyne.NewPos(abs.X, abs.Y+b.Size().Height+2))
	b.popup.Show()
}

func (b *iconTooltipButton) hideTooltip() {
	if b.popup != nil {
		b.popup.Hide()
	}
}
