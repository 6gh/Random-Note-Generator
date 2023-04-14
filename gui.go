package main

import (
	"errors"
	"fmt"
	"image/color"
	"net/url"
	"path"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func createGUI() {
	app := app.NewWithID("xyz.6gh.randomnotegen")

	window := app.NewWindow("Random Note Generator")
	window.Resize(fyne.NewSize(800, 600))

	// help bar at the bottom
	HelpBar := widget.NewToolbar(
		widget.NewToolbarAction(theme.HelpIcon(), func() {
			icon := canvas.NewImageFromResource(window.Icon())
			icon.SetMinSize(fyne.NewSize(128, 128))
			icon.FillMode = canvas.ImageFillContain

			title := canvas.NewText("Random Note Generator", color.White)
			title.TextStyle = fyne.TextStyle{
				Bold: true,
			}
			title.Alignment = fyne.TextAlignCenter
			title.TextSize = 24

			version := canvas.NewText(fyne.CurrentApp().Metadata().Version, color.White)
			version.Alignment = fyne.TextAlignCenter
			version.TextSize = 16

			creator := canvas.NewText("Created by 6gh", color.White)
			creator.Alignment = fyne.TextAlignCenter

			fyneUrl, err := url.Parse("https://fyne.io/")
			handleErr(err)
			fyneLbl := widget.NewHyperlink("Made with Fyne", fyneUrl)
			fyneLbl.Alignment = fyne.TextAlignCenter

			repoUrl, err := url.Parse("https://github.com/6gh/Random-Note-Generator")
			handleErr(err)
			repoLbl := widget.NewHyperlink("Check on GitHub", repoUrl)
			repoLbl.Alignment = fyne.TextAlignCenter

			hBox := container.New(layout.NewGridLayout(2), fyneLbl, repoLbl)
			vBox := container.New(layout.NewVBoxLayout(), icon, title, version, creator, hBox)

			dialog.ShowCustom("About", "Close", vBox, window)
		}),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			title := canvas.NewText("Additional Settings", color.White)
			title.TextStyle = fyne.TextStyle{
				Bold: true,
			}
			title.Alignment = fyne.TextAlignCenter
			title.TextSize = 24

			// max notes per track
			MaxNotesNumInput := createNumberInput(0, -1)

			// length type
			// ticks, seconds, bars
			LengthSelectInput := widget.NewSelect([]string{"MIDI Ticks", "MIDI Bars"}, func(string) {})

			// whether to cut off notes that are longer than the length of the midi
			TrimNotesChkInput := widget.NewCheck("Cut Notes", func(bool) {})

			// turn into form FormItems
			FormItems := []*widget.FormItem{
				widget.NewFormItem("Max Notes Per Track", MaxNotesNumInput),
				widget.NewFormItem("Length Type", LengthSelectInput),
				widget.NewFormItem("Trim Notes", TrimNotesChkInput),
			}

			// set default values
			MaxNotesNumInput.SetText(app.Preferences().StringWithFallback("maxNotesPerTrack", "1000"))
			LengthSelectInput.SetSelected(app.Preferences().StringWithFallback("lengthType", "MIDI Ticks"))
			TrimNotesChkInput.SetChecked(app.Preferences().BoolWithFallback("trimNotes", true))

			dialog.ShowForm("Settings", "Save", "Cancel", FormItems, func(b bool) {
				if !b {
					return
				}

				// save values
				app.Preferences().SetString("maxNotesPerTrack", MaxNotesNumInput.Text)
				app.Preferences().SetString("lengthType", LengthSelectInput.Selected)
				app.Preferences().SetBool("trimNotes", TrimNotesChkInput.Checked)
			}, window)
		}),
	)

	// 1st row
	// hosts output file
	OutputPathTxtInput := widget.NewEntry()
	OutputPathTxtLbl := widget.NewButton("Output", func() {
		// create file dialog
		// user can select a file to save to
		fileDialog := dialog.NewFileSave(func(reader fyne.URIWriteCloser, _ error) {
			if reader != nil { // if the user successfully selected a file
				filePath := reader.URI().Path()

				if path.Ext(filePath) != ".mid" {
					filePath += ".mid" // add .mid extension if it doesn't contain it
				}
				OutputPathTxtInput.SetText(filePath)
			}
		}, window)

		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".mid"}))
		fileDialog.Show()
	})
	OutputPathTxtLbl.Icon = theme.FileIcon()

	// 2nd row
	// hosts ppq, and bpm
	PPQSelectLbl := createTxt("PPQ:")
	PPQSelectInput := widget.NewSelect([]string{"96", "192", "240", "480", "960", "1920", "3840", "8192"}, func(string) {})

	BPMNumLbl := createTxt("BPM:")
	BPMNumInput := createNumberInput(0, 1000)

	// 3rd row
	// hosts ticks, NotesPerTrack
	TicksNumLbl := createTxt("MIDI Length:")
	TicksNumInput := createNumberInput(0, -1)

	NotesNumLbl := createTxt("Notes:")
	NotesNumInput := createNumberInput(0, -1)

	// 4th row
	// hosts min and max note length
	MinNoteLenNumLbl := createTxt("Min Note Length:")
	MinNoteLenNumInput := createNumberInput(0, -1)

	MaxNoteLenNumLbl := createTxt("Max Note Length:")
	MaxNoteLenNuminput := createNumberInput(0, -1)

	// output box
	OutputLogTxt := widget.NewMultiLineEntry()
	OutputLogTxt.SetText("Output will go here...")

	// create button
	CreateBTN := widget.NewButton("Create", func() {
		var errors []string

		// validate all inputs
		// if any are invalid, add them to the error list
		// note: there has got to be a better way to do this lol
		if err := OutputPathTxtInput.Validate(); err != nil {
			errors = append(errors, "output: "+err.Error())
		}
		if PPQSelectInput.Selected == "" {
			errors = append(errors, "ppq: cannot be empty")
		}
		if err := BPMNumInput.Validate(); err != nil {
			errors = append(errors, "bpm: "+err.Error())
		}
		if err := TicksNumInput.Validate(); err != nil {
			errors = append(errors, "ticks: "+err.Error())
		}
		if err := NotesNumInput.Validate(); err != nil {
			errors = append(errors, "notes: "+err.Error())
		}

		if len(errors) > 0 {
			// if there are any errors show them in a dialog, and do not continue
			dialog.ShowInformation("Invalid Options", strings.Join(errors, "\n"), window)
		} else {
			// if there are no errors, create the midi file
			OutputLogTxt.SetText("")

			// get values from inputs, converting to correct types
			noteCount, err := strconv.Atoi(NotesNumInput.Text)
			handleErr(err)
			ticks, err := strconv.Atoi(TicksNumInput.Text)
			handleErr(err)
			minNoteLength, err := strconv.Atoi(MinNoteLenNumInput.Text)
			handleErr(err)
			maxNoteLength, err := strconv.Atoi(MaxNoteLenNuminput.Text)
			handleErr(err)
			maxNotesPerTrack, err := strconv.Atoi(app.Preferences().StringWithFallback("maxNotesPerTrack", "1000"))
			handleErr(err)
			ppq, err := strconv.Atoi(PPQSelectInput.Selected)
			handleErr(err)
			bpm, err := strconv.Atoi(BPMNumInput.Text)
			handleErr(err)
			trimNotes := app.Preferences().BoolWithFallback("trimNotes", true)

			// if user selected MIDI Bars, convert the bars to ticks
			lengthType := app.Preferences().StringWithFallback("lengthType", "MIDI Ticks")

			if lengthType == "MIDI Bars" {
				// ticks rn is the number of bars
				// so we need to convert it to ticks, by multiplying it by the ppq
				// ppq is the number of ticks per quarter note, so we need to multiply it by 4
				// ticks = bars * ppq * 4
				ticks = int(float64(ticks) * float64(ppq) * 4)
			}

			// disable all inputs
			OutputPathTxtInput.Disable()
			TicksNumInput.Disable()
			MinNoteLenNumInput.Disable()
			MaxNoteLenNuminput.Disable()
			PPQSelectInput.Disable()
			BPMNumInput.Disable()
			window.SetTitle("Random Note Generator (Running...)")
			// TODO: add a cancel button
			// TODO: disable the create button

			// log the values
			OutputLogTxt.SetText(
				fmt.Sprintf(
					"creating tracks | nc: %d | len: %d | maxlen: %d | minlen: %d | notesper: %d | trimnotes: %t\n",
					noteCount,
					ticks,
					maxNoteLength,
					minNoteLength,
					maxNotesPerTrack,
					trimNotes,
				),
			)

			// create the tracks
			tracks := createTracks(
				noteCount,
				ticks,
				maxNoteLength,
				minNoteLength,
				maxNotesPerTrack,
				trimNotes,
				func(format string, args ...any) {
					OutputLogTxt.SetText(OutputLogTxt.Text + fmt.Sprintf(format, args...) + "\n")
				},
			)
			OutputLogTxt.SetText(OutputLogTxt.Text + "created tracks" + "\n")

			// save the tracks to a midi file
			OutputLogTxt.SetText(OutputLogTxt.Text + "saving to midi" + "\n")
			createMIDI(OutputPathTxtInput.Text, ppq, bpm, tracks, func() {
				OutputLogTxt.SetText(OutputLogTxt.Text + "saved to midi" + "\n")

				// after the midi file is saved, enable all inputs
				OutputPathTxtInput.Enable()
				TicksNumInput.Enable()
				MinNoteLenNumInput.Enable()
				MaxNoteLenNuminput.Enable()
				PPQSelectInput.Enable()
				BPMNumInput.Enable()
				window.SetTitle("Random Note Generator")
			})
		}
	})

	// set default values
	// or values from saved preferences
	OutputPathTxtInput.SetText(app.Preferences().StringWithFallback("outputPath", "output.mid"))

	PPQSelectInput.SetSelected(app.Preferences().StringWithFallback("ppq", "960"))
	BPMNumInput.SetText(app.Preferences().StringWithFallback("bpm", "120"))

	TicksNumInput.SetText(app.Preferences().StringWithFallback("ticks", "122880"))
	NotesNumInput.SetText(app.Preferences().StringWithFallback("notes", "20000"))

	MinNoteLenNumInput.SetText(app.Preferences().StringWithFallback("minNoteLength", "960"))
	MaxNoteLenNuminput.SetText(app.Preferences().StringWithFallback("maxNoteLength", "1920"))

	// create content container
	content := container.NewBorder(
		container.NewVBox(
			container.New( // outputbtn outputtxt
				layout.NewFormLayout(),
				OutputPathTxtLbl,
				OutputPathTxtInput,
			),
			container.New( // ppqlbl ppqtxt | bpmlbl bpmtxt
				layout.NewGridLayout(2),
				container.New(layout.NewFormLayout(), PPQSelectLbl, PPQSelectInput),
				container.New(layout.NewFormLayout(), BPMNumLbl, BPMNumInput),
			),
			container.New( // tickslbl tickstxt | noteslbl notestxt
				layout.NewGridLayout(2),
				container.New(layout.NewFormLayout(), TicksNumLbl, TicksNumInput),
				container.New(layout.NewFormLayout(), NotesNumLbl, NotesNumInput),
			),
			container.New( // minnotelbl minnotetxt | maxnotelbl maxnotetxt
				layout.NewGridLayout(2),
				container.New(layout.NewFormLayout(), MinNoteLenNumLbl, MinNoteLenNumInput),
				container.New(layout.NewFormLayout(), MaxNoteLenNumLbl, MaxNoteLenNuminput),
			),
			CreateBTN,
		),
		HelpBar,
		nil,
		nil,
		container.New(
			layout.NewMaxLayout(),
			container.New(
				layout.NewMaxLayout(),
				OutputLogTxt,
			),
		),
	)

	// set close intercept to save settings
	window.SetCloseIntercept(func() {
		logf("GUI closed, saving settings")

		app.Preferences().SetString("outputPath", OutputPathTxtInput.Text)
		app.Preferences().SetString("ppq", PPQSelectInput.Selected)
		app.Preferences().SetString("bpm", BPMNumInput.Text)
		app.Preferences().SetString("ticks", TicksNumInput.Text)
		app.Preferences().SetString("notes", NotesNumInput.Text)

		window.Close()
	})

	// set content and show
	window.SetContent(content)
	window.ShowAndRun()
}

// Helper function to create Text Inputs with the same settings
func createTxt(text string) *canvas.Text {
	txt := canvas.NewText(text, color.White)
	txt.Alignment = fyne.TextAlignLeading
	txt.TextSize = 16

	return txt
}

// Helper function to create Number Inputs with the same settings
// If max is -1, there is no max
func createNumberInput(min int, max int) *widget.Entry {
	entry := widget.NewEntry()
	entry.Validator = func(input string) error {
		if input == "" {
			return errors.New("cannot be empty")
		}

		num, err := strconv.Atoi(input)

		if err != nil {
			return errors.New("not a number")
		}

		if num < min {
			return errors.New("number too small")
		}

		if max != -1 && num > max {
			return errors.New("number too large")
		}

		return nil
	}
	return entry
}
