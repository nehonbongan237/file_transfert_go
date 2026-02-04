package gui

import (
	"file_transfert_go/client/network"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ShowLogin(a fyne.App) {
	w := a.NewWindow("Connexion")

	username := widget.NewEntry()
	password := widget.NewPasswordEntry()
	status := widget.NewLabel("")

	btn := widget.NewButton("Connexion", func() {
		token := network.Authenticate(username.Text, password.Text)
		if token != "" {
			ShowFiles(a)
			w.Close()
		} else {
			status.SetText("❌ Authentification échouée")
		}
	})

	w.SetContent(container.NewVBox(
		widget.NewLabel("Utilisateur"),
		username,
		widget.NewLabel("Mot de passe"),
		password,
		btn,
		status,
	))

	w.Resize(fyne.NewSize(300, 250))
	w.Show()
}
