package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

func main() {
	a := app.NewWithID("com.github.elecshan.crocfyne")
	w := a.NewWindow("crocfyne")

	w.SetContent(container.NewBorder(nil, nil, nil, nil,
		container.NewAppTabs(
			sendTabItem(a, w),
			receiveTabItem(a, w))))

	a.Preferences().SetString("relayAddress", a.Preferences().StringWithFallback("relayAddress", "croc.schollz.com:9009"))
	a.Preferences().SetString("relayPorts", a.Preferences().StringWithFallback("relayPorts", "9009,9010,9011,9012,9013"))
	a.Preferences().SetString("relayPassword", a.Preferences().StringWithFallback("relayPassword", "pass123"))
	a.Preferences().SetBool("disableLocal", a.Preferences().BoolWithFallback("disableLocal", false))
	a.Preferences().SetBool("noMultiplexing", a.Preferences().BoolWithFallback("noMultiplexing", false))
	a.Preferences().SetBool("forceLocal", a.Preferences().BoolWithFallback("forceLocal", false))
	a.Preferences().SetBool("disableCompression", a.Preferences().BoolWithFallback("disableCompression", false))
	a.Preferences().SetString("pakeCurve", a.Preferences().StringWithFallback("pakeCurve", "siec"))

	w.Resize(fyne.NewSize(600, 400))
	w.ShowAndRun()
}
