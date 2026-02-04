package gui

import (
	"file_transfert_go/client/network"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// Struct pour gérer l’interface
type FileManager struct {
	window       fyne.Window
	fileList     []string
	selectedFile string
	groupSelect  *widget.Select
	filesList    *widget.List
	statusLabel  *widget.Label
	progressBar  *widget.ProgressBar
}

// Constructeur
func NewFileManager(a fyne.App) *FileManager {
	return &FileManager{
		window:   a.NewWindow("Gestionnaire de Fichiers - Serveur"),
		fileList: make([]string, 0),
	}
}

// Initialisation de l’UI
func (fm *FileManager) setupUI() {
	fm.statusLabel = widget.NewLabel("Sélectionnez un groupe puis actualisez")
	fm.statusLabel.Alignment = fyne.TextAlignCenter

	// Liste des fichiers
	fm.filesList = widget.NewList(
		func() int { return len(fm.fileList) },
		func() fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(nil),
				widget.NewLabel(""),
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			box := o.(*fyne.Container)
			label := box.Objects[1].(*widget.Label)
			label.SetText(fm.fileList[i])
		},
	)
	fm.filesList.OnSelected = fm.onFileSelected
	fm.filesList.OnUnselected = fm.onFileUnselected

	// Sélecteur de groupe
	userGroups, err := network.GetUserGroups()
if err != nil || len(userGroups) == 0 {
    dialog.ShowError(fmt.Errorf("Impossible de récupérer vos groupes"), fm.window)
    return
}

fm.groupSelect = widget.NewSelect(userGroups, fm.onGroupChanged)
fm.groupSelect.Selected = userGroups[0] // sélectionner le premier automatiquement

// Si l'utilisateur n'a qu'un seul groupe, désactiver la sélection
if len(userGroups) == 1 {
    fm.groupSelect.Disable()
}

	// Boutons
	refreshBtn := widget.NewButton("Actualiser", fm.refreshFiles)
	downloadBtn := widget.NewButton("Télécharger", fm.downloadFile)
	uploadBtn := widget.NewButton("Téléverser", fm.uploadFile)

	// ProgressBar
	fm.progressBar = widget.NewProgressBar()
	fm.progressBar.Hide()

	// Container principal
	content := container.NewVBox(
		fm.groupSelect,
		fm.statusLabel,
		fm.filesList,
		container.NewHBox(refreshBtn, downloadBtn, uploadBtn),
		fm.progressBar,
	)

	fm.window.SetContent(content)
	fm.window.Resize(fyne.NewSize(600, 400))
}

// Changement de groupe
func (fm *FileManager) onGroupChanged(selected string) {
	if selected != "" {
		fm.statusLabel.SetText("Groupe sélectionné: " + selected + " - Cliquez sur Actualiser")
		fm.fileList = make([]string, 0)
		fm.selectedFile = ""
		fm.filesList.Refresh()
	}
}

// Sélection / désélection d’un fichier
func (fm *FileManager) onFileSelected(id widget.ListItemID) {
	if id < len(fm.fileList) {
		fm.selectedFile = fm.fileList[id]
		fm.statusLabel.SetText("Fichier sélectionné: " + fm.selectedFile)
	}
}

func (fm *FileManager) onFileUnselected(id widget.ListItemID) {
	fm.selectedFile = ""
	fm.statusLabel.SetText("Aucun fichier sélectionné")
}

// Actualiser la liste des fichiers
func (fm *FileManager) refreshFiles() {
	if fm.groupSelect.Selected == "" {
		dialog.ShowInformation("Attention", "Veuillez d'abord sélectionner un groupe", fm.window)
		return
	}

	fm.progressBar.Show()
	fm.statusLabel.SetText("Chargement des fichiers...")

	go func() {
		newFileList := network.ListFiles(fm.groupSelect.Selected)

		// Mise à jour UI
		fm.window.Canvas().Refresh(fm.filesList)
		fm.fileList = newFileList
		fm.filesList.Refresh()
		fm.progressBar.Hide()

		if len(newFileList) == 0 {
			fm.statusLabel.SetText("Aucun fichier trouvé")
		} else {
			fm.statusLabel.SetText(fmt.Sprintf("%d fichier(s) trouvé(s)", len(newFileList)))
		}
	}()
}

// Télécharger un fichier
func (fm *FileManager) downloadFile() {
	if fm.groupSelect.Selected == "" {
		dialog.ShowInformation("Erreur", "Veuillez sélectionner un groupe", fm.window)
		return
	}
	if fm.selectedFile == "" {
		dialog.ShowInformation("Erreur", "Veuillez sélectionner un fichier", fm.window)
		return
	}

	dialog.ShowConfirm(
		"Confirmation",
		fmt.Sprintf("Télécharger le fichier '%s' ?", fm.selectedFile),
		func(confirmed bool) {
			if confirmed {
				fm.performDownload()
			}
		},
		fm.window,
	)
}

func (fm *FileManager) performDownload() {
	fm.progressBar.Show()
	fm.statusLabel.SetText("Téléchargement en cours...")

	go func() {
		err := network.Download(fm.groupSelect.Selected, fm.selectedFile)
		fm.progressBar.Hide()

		if err != nil {
			dialog.ShowError(err, fm.window)
			fm.statusLabel.SetText("Erreur lors du téléchargement")
			return
		}

		dialog.ShowInformation("Succès",
			fmt.Sprintf("'%s' téléchargé avec succès", fm.selectedFile),
			fm.window)
		fm.statusLabel.SetText("Téléchargement terminé")
	}()
}

// Téléverser un fichier
func (fm *FileManager) uploadFile() {
	if fm.groupSelect.Selected == "" {
		dialog.ShowInformation("Erreur", "Veuillez sélectionner un groupe", fm.window)
		return
	}

	dialog.ShowFileOpen(
		func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				dialog.ShowError(err, fm.window)
				return
			}
			if reader == nil {
				return
			}

			dialog.ShowConfirm(
				"Confirmation",
				fmt.Sprintf("Téléverser le fichier '%s' ?", reader.URI().Name()),
				func(confirmed bool) {
					if confirmed {
						fm.performUpload(reader)
					}
				},
				fm.window,
			)
		},
		fm.window,
	)
}

func (fm *FileManager) performUpload(file fyne.URIReadCloser) {
	defer file.Close()
	fm.progressBar.Show()
	fm.statusLabel.SetText("Téléversement en cours...")

	go func() {
		err := network.Upload(fm.groupSelect.Selected, file)
		fm.progressBar.Hide()

		if err != nil {
			dialog.ShowError(err, fm.window)
			fm.statusLabel.SetText("Erreur lors du téléversement")
			return
		}

		dialog.ShowInformation("Succès",
			fmt.Sprintf("'%s' téléversé avec succès", file.URI().Name()),
			fm.window)
		fm.statusLabel.SetText("Téléversement terminé")

		// Rafraîchir la liste après upload
		fm.refreshFiles()
	}()
}

// Fonction pour lancer l’UI
func ShowFiles(a fyne.App) {
	fileManager := NewFileManager(a)
	fileManager.setupUI()
	fileManager.window.Show()
}
