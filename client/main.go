package main

import (
	"file_transfert_go/client/gui"
	"file_transfert_go/client/network"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2"
)

func main() {
	myApp := app.New()
	win := myApp.NewWindow("Client FTP")

	status := widget.NewLabel("Recherche du serveur...")
	win.SetContent(container.NewCenter(status))
	win.Resize(fyne.NewSize(400, 200))
	win.Show()

	// Canal pour notifier le main thread
	connected := make(chan error)

	// 🔹 THREAD DE CONNEXION
	go func() {
		err := network.Connect()
		connected <- err // envoie le résultat au thread principal
	}()

	// Boucle principale : on attend le résultat de la connexion
	go func() {
		err := <-connected
		if err != nil {
			status.SetText("Erreur : " + err.Error())
			return
		}
		// Afficher la page login après connexion
		gui.ShowLogin(myApp)
	}()

	myApp.Run()
}
