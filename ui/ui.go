package ui

import (
	"fmt"
	"log"
	"os"

	"github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

var p *widgets.Gauge
var l *widgets.Paragraph

func Start() {
	if err := termui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}

	l = widgets.NewParagraph()
	l.SetRect(0, 0, 50, 5)

	p = widgets.NewGauge()
	p.Title = "Progress"
	p.SetRect(0, 5, 50, 10)
	p.Percent = 0
	p.BarColor = termui.ColorBlue

	termui.Render(l, p)
}

func Stop() {
	termui.Close()
}

func UpdateProgress(percent int) {
	p.Percent = percent
	if percent == 100 {
		p.BarColor = termui.ColorGreen
	}
	termui.Render(p)
}

func PrintMessage(message string) {
	l.Text = message
	termui.Render(l)
}

func PrintError(err error) {
	l.Text = fmt.Sprintf("[error] %v", err)
	l.TextStyle.Fg = termui.ColorRed
	termui.Render(l)
	os.Exit(1)
}
