package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/schollz/croc/v9/src/croc"
	"github.com/schollz/croc/v9/src/utils"
	"github.com/schollz/logger"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func sendTabItem(a fyne.App, w fyne.Window) *container.TabItem {
	fileLabel := widget.NewLabel("Pick a file to send")
	// filePath := binding.NewString()
	// fileEntry := widget.NewEntryWithData(filePath)

	boxHolder := container.NewVBox()
	// fileData := binding.BindStringList(&[]string{})
	// fileList := widget.NewListWithData(fileData,
	// 	func() fyne.CanvasObject {
	// 		// return widget.NewLabel("template")
	// 		form := &widget.Form{Items: []*widget.FormItem{{Text: "", Widget: widget.NewLabel("text string")}}}
	// 		return form
	// 	},
	// 	func(i binding.DataItem, o fyne.CanvasObject) {
	// 		o.(*widget.Label).Bind(i.(binding.String))
	// 	})

	code := binding.NewString()
	codeEntry := widget.NewEntryWithData(code)
	codeForm := widget.NewForm(&widget.FormItem{Text: "Send code", Widget: codeEntry})

	progBar := widget.NewProgressBar()
	progBar.Hide()

	sendDir, _ := os.MkdirTemp("", "croc-send")
	randomCode := utils.GetRandomName()
	code.Set(randomCode)

	status := binding.NewString()
	statusLabel := widget.NewLabelWithData(status)
	statusLabel.Hide()

	fileEntries := make(map[string]*fyne.Container)
	var fileButton, sendButton, cancelButton *widget.Button
	// showWidgetHolder := container.NewVBox()
	cancelChan := make(chan bool)

	cancelButton = widget.NewButtonWithIcon("Cancel", theme.CancelIcon(), func() {
		cancelChan <- true
	})
	cancelButton.Disable()

	fileButton = widget.NewButtonWithIcon("Select", theme.FileIcon(), func() {
		dialog.ShowFileOpen(func(f fyne.URIReadCloser, e error) {
			if e != nil {
				logger.Errorf("Open file dialog error: %s", e.Error())
				return
			}
			if f != nil {
				tfile, oerr := os.Create(filepath.Join(sendDir, f.URI().Name()))
				if oerr != nil {
					logger.Errorf("Unable to create temp file: %s --%s", tfile.Name(), oerr.Error())
					return
				}
				io.Copy(tfile, f)
				tfile.Close()
				fpath := tfile.Name()
				logger.Tracef("Android URL (%s), copied to internal cache (%s)", f.URI().Name(), fpath)

				// filePath.Set(fpath)
				// fileData.Append(fpath)
				newEntry := container.NewHBox(widget.NewLabel(filepath.Base(fpath)), layout.NewSpacer(), widget.NewButtonWithIcon("", theme.CancelIcon(), func() {
					if !sendButton.Disabled() {
						if fe, ok := fileEntries[fpath]; ok {
							boxHolder.Remove(fe)
							os.Remove(fpath)
							logger.Tracef("Removed file from internal cache: %s", fpath)
							delete(fileEntries, fpath)
						}
					}
				}))
				fileEntries[fpath] = newEntry
				boxHolder.Add(newEntry)
			}
		}, w)
	})

	resetSender := func() {
		progBar.Hide()
		progBar.SetValue(0)

		for file, entry := range fileEntries {
			boxHolder.Remove(entry)
			os.Remove(file)
			logger.Tracef("Remove file from internal cache: %s", file)
			delete(fileEntries, file)
		}

		if codeEntry.Text == randomCode {
			randomCode = utils.GetRandomName()
			code.Set(randomCode)
		}
		codeEntry.Enable()
		fileButton.Enable()
		sendButton.Enable()
		cancelButton.Disable()
	}

	sendButton = widget.NewButtonWithIcon("Send", theme.MailSendIcon(), func() {
		if len(fileEntries) < 1 {
			logger.Error("No file to send")
			return
		}

		fileButton.Disable()
		status.Set("")
		sender, err := croc.New(croc.Options{
			IsSender:       true,
			SharedSecret:   randomCode,
			Debug:          false,
			RelayAddress:   a.Preferences().String("relayAddress"),
			RelayPorts:     strings.Split(a.Preferences().String("relayPorts"), ","),
			RelayPassword:  a.Preferences().String("relayPassword"),
			Stdout:         false,
			NoPrompt:       true,
			DisableLocal:   a.Preferences().Bool("disableLocal"),
			NoMultiplexing: a.Preferences().Bool("noMultiplexing"),
			OnlyLocal:      a.Preferences().Bool("forceLocal"),
			NoCompress:     a.Preferences().Bool("disableCompression"),
			Curve:          a.Preferences().String("pakeCurve"),
		})

		if err != nil {
			logger.Errorf("croc error: %s\n", err.Error())
			return
		}
		logger.SetLevel("croc")
		logger.Trace("croc sender created")

		doneChan := make(chan bool)

		go func() {
			ticker := time.NewTicker(time.Microsecond * 100)
			for {
				select {
				case <-ticker.C:
					if sender.Step2FileInfoTransfered {
						num := sender.FilesToTransferCurrentNum
						file := sender.FilesToTransfer[num]
						fname := filepath.Base(file.Name)
						progBar.Max = float64(file.Size)
						progBar.SetValue(float64(sender.TotalSent))
						status.Set("Send file " + fname)
					}
				case <-doneChan:
					ticker.Stop()
					return
				}
			}
		}()

		go func() {
			codeEntry.Disable()
			progBar.Show()
			cancelButton.Enable()
			sendButton.Disable()
			statusLabel.Show()

			files := []string{}
			for f := range fileEntries {
				files = append(files, f)
			}

			serr := sender.Send(croc.TransferOptions{PathToFiles: files})

			doneChan <- true

			if serr != nil {
				logger.Errorf("Send files failed: %s", serr.Error())
			} else {
				status.Set("Send files success!")
			}
			resetSender()
		}()

		go func() {
			<-cancelChan
			doneChan <- true
			status.Set("Send cancelled.")
			resetSender()
		}()
	})

	return container.NewTabItemWithIcon("send", theme.MailSendIcon(),
		container.NewVBox(container.NewHBox(fileLabel, fileButton),
			codeForm,
			sendButton,
			cancelButton,
			progBar,
			boxHolder,
			statusLabel))
}
