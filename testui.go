package main

import (
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Test UI")

	label := widget.NewLabel("Si tu vois ça, Fyne fonctionne !")
	w.SetContent(label)

	w.ShowAndRun()
}
