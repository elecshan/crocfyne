package main

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func receiveTabItem() *container.TabItem {
	tabItem := container.NewTabItemWithIcon("receive", theme.DownloadIcon(),
		container.NewVBox(widget.NewLabel("text string")))

	return tabItem
}
