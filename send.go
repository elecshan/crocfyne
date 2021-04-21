package main

import (
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func sendTabItem() *container.TabItem {
	tabItem := container.NewTabItemWithIcon("send", theme.MailSendIcon(),
		container.NewVBox(widget.NewLabel("text string")))

	return tabItem
}
