package main

import (
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/schollz/croc/v9/src/croc"
	"github.com/schollz/logger"
)

func receiveTabItem(a fyne.App, w fyne.Window) *container.TabItem {
	fileLabel := widget.NewLabel("Enter code to download")

	code := binding.NewString()
	codeEntry := widget.NewEntryWithData(code)
	codeForm := widget.NewForm(&widget.FormItem{Text: "Send code", Widget: codeEntry})

	progBar := widget.NewProgressBar()
	progBar.Hide()

	status := binding.NewString()
	statusLabel := widget.NewLabelWithData(status)
	statusLabel.Hide()

	recvDir := binding.NewString()
	pathEntry := widget.NewEntryWithData(recvDir)
	pathButton := widget.NewButtonWithIcon("Select", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(url fyne.ListableURI, e error) {
			if e != nil {
				logger.Errorf("Open folder dialog error: %s", e.Error())
				return
			}
			if url != nil {
				logger.Info(url.Path())
				recvDir.Set(url.Path())
			}
		}, w)
	})
	pathContainer := container.NewHBox(pathEntry, pathButton)
	pathForm := widget.NewForm(&widget.FormItem{Text: "Select a folder", Widget: pathContainer})

	var recvButton, cancelButton *widget.Button

	cancelChan := make(chan bool)

	cancelButton = widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		cancelChan <- true
	})
	cancelButton.Disable()

	// recvDir, _ := os.MkdirTemp("", "croc-recv")

	recvButton = widget.NewButtonWithIcon("Download", theme.DownloadIcon(), func() {
		receiver, err := croc.New(croc.Options{
			IsSender:       false,
			SharedSecret:   codeEntry.Text,
			Debug:          false,
			RelayAddress:   a.Preferences().String("relay-address"),
			RelayPassword:  a.Preferences().String("relay-password"),
			Stdout:         false,
			NoPrompt:       true,
			DisableLocal:   a.Preferences().Bool("disable-local"),
			NoMultiplexing: a.Preferences().Bool("disable-multiplexing"),
			OnlyLocal:      a.Preferences().Bool("force-local"),
			NoCompress:     a.Preferences().Bool("disable-compression"),
			Curve:          a.Preferences().String("pake-curve"),
		})
		if err != nil {
			logger.Errorf("Receive setup error: %s", err.Error())
			return
		}

		logger.Trace("croc receiver created")

		doneChan := make(chan bool)

		go func() {
			ticker := time.NewTicker(time.Microsecond * 100)
			for {
				select {
				case <-ticker.C:
					if receiver.Step2FileInfoTransfered {
						num := receiver.FilesToTransferCurrentNum
						file := receiver.FilesToTransfer[num]
						fname := filepath.Base(file.Name)
						progBar.Max = float64(file.Size)
						progBar.SetValue(float64(receiver.TotalSent))
						status.Set("Download file " + fname)
					}
				case <-doneChan:
					ticker.Stop()
					return
				}
			}
		}()

		err = os.Chdir(pathEntry.Text)
		if err != nil {
			logger.Errorf("Change directory error: %s", err.Error())
		}
		status.Set("")
		err = receiver.Receive()
		doneChan <- true
		progBar.SetValue(0)
		progBar.Hide()

		if err != nil {
			logger.Errorf("Download file failed: %s", err.Error())
		} else {
			// filepath.WalkDir(recvDir, func(path string, info fs.DirEntry, err error) error {
			// 	if !info.IsDir() {

			// 	}
			// })
		}
	})

	return container.NewTabItemWithIcon("receive", theme.DownloadIcon(),
		container.NewVBox(fileLabel,
			codeForm,
			pathForm,
			recvButton,
			cancelButton,
			progBar,
			statusLabel))
}
