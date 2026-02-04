package gui

import (
	"file_transfert_go/serveur/storage"
	"file_transfert_go/serveur/users"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func ServerUI(w fyne.Window) fyne.CanvasObject {

	/************* COLONNE GAUCHE : FICHIERS *************/

	var fileList []string
	selectedFile := ""

	files := widget.NewList(
		func() int { return len(fileList) },
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(fileList[i])
		},
	)

	files.OnSelected = func(id widget.ListItemID) {
		selectedFile = fileList[id]
	}

	btnRefresh := widget.NewButton("🔄 Rafraîchir", func() {
		var err error
		fileList, err = storage.ListFiles("default")
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		files.Refresh()
	})

	leftColumn := container.NewVBox(
		widget.NewLabelWithStyle(
			"📁 Fichiers sur le serveur",
			fyne.TextAlignCenter,
			fyne.TextStyle{Bold: true},
		),
		files,
		btnRefresh,
	)

	/************* COLONNE DROITE : ADMIN *************/

	/* --- Téléversement --- */

	groupEntry := widget.NewEntry()
	groupEntry.SetPlaceHolder("Nom du groupe")

	btnUpload := widget.NewButton("📤 Téléverser un fichier", func() {

		dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil || r == nil {
				return
			}

			err = storage.SaveFile(r.URI().Path(), groupEntry.Text)
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			dialog.ShowInformation("Succès", "Fichier téléversé", w)
			btnRefresh.OnTapped()
		}, w)
	})

	uploadBox := container.NewVBox(
		widget.NewLabelWithStyle("Téléverser un fichier",
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		),
		groupEntry,
		btnUpload,
	)

	/* --- Création utilisateur --- */

	username := widget.NewEntry()
	password := widget.NewPasswordEntry()
	userGroup := widget.NewEntry()

	username.SetPlaceHolder("Nom d'utilisateur")
	password.SetPlaceHolder("Mot de passe")
	userGroup.SetPlaceHolder("Groupe")

	btnCreateUser := widget.NewButton("👤 Créer utilisateur", func() {

		if username.Text == "" || password.Text == "" || userGroup.Text == "" {
			dialog.ShowInformation("Erreur", "Tous les champs sont requis", w)
			return
		}

		err := users.CreateUser(
			username.Text,
			password.Text,
			userGroup.Text,
		)

		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		dialog.ShowInformation("Succès", "Utilisateur créé", w)
		username.SetText("")
		password.SetText("")
		userGroup.SetText("")
	})

	userBox := container.NewVBox(
		widget.NewLabelWithStyle(
			"Créer un utilisateur",
			fyne.TextAlignLeading,
			fyne.TextStyle{Bold: true},
		),
		username,
		password,
		userGroup,
		btnCreateUser,
	)

	rightColumn := container.NewVBox(
		uploadBox,
		widget.NewSeparator(),
		userBox,
	)

	/************* LAYOUT FINAL *************/

	return container.NewGridWithColumns(
		2,
		leftColumn,
		rightColumn,
	)
}
