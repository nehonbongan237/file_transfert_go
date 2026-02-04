package main

import (
	"file_transfert_go/client/gui"
	"file_transfert_go/client/network"
	"log"

	"fyne.io/fyne/v2/app"
)

func main() {
	err := network.Connect()
	if err != nil {
		log.Fatal("Impossible de se connecter au serveur")
	}

	a := app.New()
	println("Avant ShowAndRun")
	gui.ShowLogin(a)
	a.Run()
}
