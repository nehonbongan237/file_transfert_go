package gui

import (
	
	"file_transfert_go/serveur/groupes"
	"file_transfert_go/serveur/storage"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

func ShowAdminUI(app fyne.App) {
	win := app.NewWindow("Admin Panel")
	win.Resize(fyne.NewSize(900, 500))

	/* =======================
	   COLONNE UTILISATEURS
	   ======================= */

	userList := widget.NewList(
    func() int {
        return len(storage.GetAllUsers()) // On appelle à chaque rendu
    },
    func() fyne.CanvasObject {
        return widget.NewLabel("")
    },
    func(i widget.ListItemID, o fyne.CanvasObject) {
        users := storage.GetAllUsers() // On récupère à chaque fois
        o.(*widget.Label).SetText(users[i])
    },
	)


	// Suppression utilisateur
	userList.OnSelected = func(id widget.ListItemID) {
		 users := storage.GetAllUsers()
		user := users[id]

		dialog.ShowConfirm(
			"Supprimer utilisateur",
			"Supprimer l'utilisateur "+user+" ?",
			func(ok bool) {
				if ok {
					storage.DeleteUser(user) // à créer
					users = storage.GetAllUsers()
					userList.Refresh()
				}
			},
			win,
		)
	
	}

	// Création utilisateur
	createBtn := widget.NewButton("➕ Create User", func() {
		username := widget.NewEntry()
		password := widget.NewPasswordEntry()
		group := widget.NewEntry()

		form := dialog.NewForm(
			"Créer utilisateur",
			"Créer",
			"Annuler",
			[]*widget.FormItem{
				widget.NewFormItem("Username", username),
				widget.NewFormItem("Password", password),
				widget.NewFormItem("Group", group),
			},
			func(ok bool) {
				if ok {
					groups.AddUserWithCredentials(
						username.Text,
						password.Text,
						group.Text,
					)
					
					userList.Refresh()
				}
			},
			win,
		)
		form.Show()
	})

	userCol := container.NewBorder(
		widget.NewLabel("👤 Utilisateurs"),
		createBtn,
		nil,
		nil,
		userList,
	)

	/* =======================
	   COLONNE FICHIERS
	   ======================= */

	files,_ := storage.ListAllFiles()
	 fileList := widget.NewList(
        func() int { return len(files) },
        func() fyne.CanvasObject { return widget.NewLabel("") },
        func(i widget.ListItemID, o fyne.CanvasObject) {
            o.(*widget.Label).SetText(files[i])
        })

	fileList.OnSelected = func(id widget.ListItemID) {
		file := files[id]

		dialog.ShowConfirm(
			"Supprimer fichier",
			"Supprimer le fichier "+file+" ?",
			func(ok bool) {
				if ok {
					storage.DeleteFile(file) // à créer
					files,_= storage.ListAllFiles()
					fileList.Refresh()
				}
			},
			win,
		)
	}

	fileCol := container.NewBorder(
		widget.NewLabel("📁 Fichiers serveur"),
		nil,
		nil,
		nil,
		fileList,
	)

	/* =======================
	   LAYOUT FINAL
	   ======================= */

	split := container.NewHSplit(userCol, fileCol)
	split.Offset = 0.45

	win.SetContent(split)
	win.Show()
}
