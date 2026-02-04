package main

import (
	"log"
	"file_transfert_go/serveur/network"
)

func main() {
	log.Println("Démarrage du serveur...")
	network.StartServer(":8080")
}
