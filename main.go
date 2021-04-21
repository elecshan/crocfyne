package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
)

func main() {
	a := app.NewWithID("com.github.elecshan.crocfyne")
	w := a.NewWindow("crocfyne")

	w.SetContent(container.NewBorder(nil, nil, nil, nil,
		container.NewAppTabs(
			sendTabItem(),
			receiveTabItem())))

	w.ShowAndRun()
}
