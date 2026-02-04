package main

import (
	"log"
	"file_transfert_go/serveur/network"
	"file_transfert_go/serveur/gui"
	"fyne.io/fyne/v2/app"
)

func main() {
	log.Println("Démarrage du serveur...")
	go network.StartServer(":8080")

	a := app.New()
	println("Avant ShowAndRun")
	gui.ShowAdminUI(a)
	a.Run()
}
 

