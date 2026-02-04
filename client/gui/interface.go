package gui

import (
	"file_transfert_go/client/network"
	"fmt"
	"strings"
	"os"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

// FileManager gère l'interface
type FileManager struct {
	window       fyne.Window
	fileList     []string
	selectedFile string
	group        string
	filesList    *widget.List
	statusLabel  *widget.Label
	progressBar  *widget.ProgressBar
}

// Création du FileManager
func NewFileManager(a fyne.App) *FileManager {
	return &FileManager{
		window:   a.NewWindow("Gestionnaire de Fichiers"),
		fileList: []string{},
	}
}

// Initialisation de l'UI
func (fm *FileManager) setupUI() {
	fm.statusLabel = widget.NewLabel("Chargement des fichiers...")
	fm.statusLabel.Alignment = fyne.TextAlignCenter

	fm.progressBar = widget.NewProgressBar()
	fm.progressBar.Hide()

	// Liste des fichiers
	fm.filesList = widget.NewList(
		func() int { return len(fm.fileList) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, o fyne.CanvasObject) {
			label := o.(*widget.Label)
			label.SetText(fm.fileList[i])
		},
	)
	fm.filesList.OnSelected = fm.onFileSelected
	fm.filesList.OnUnselected = fm.onFileUnselected

	// Boutons
	refreshBtn := widget.NewButton("Actualiser", fm.refreshFiles)
	downloadBtn := widget.NewButton("Télécharger", fm.downloadFile)
	uploadBtn := widget.NewButton("Téléverser", fm.uploadFile)
	serverHistoryBtn := widget.NewButton("Historique serveur", fm.showServerHistory)

	// Layout principal
	filesScroll := container.NewVScroll(fm.filesList)
	filesScroll.SetMinSize(fyne.NewSize(800, 300)) // largeur x hauteur

	content := container.NewVBox(
		fm.statusLabel,
		fm.progressBar,
		filesScroll,
		container.NewHBox(refreshBtn, downloadBtn, uploadBtn, serverHistoryBtn),
	)

	fm.window.SetContent(content)
	fm.window.Resize(fyne.NewSize(800, 500))

	// Récupérer le groupe utilisateur automatiquement depuis le serveur
	userGroups, err := network.GetUserGroups()
	if err != nil || len(userGroups) == 0 {
		dialog.ShowError(fmt.Errorf("Impossible de récupérer vos groupes"), fm.window)
		return
	}
	fm.group = userGroups[0] // utiliser le premier groupe par défaut
	if len(userGroups) > 1 {
		dialog.ShowInformation("Info", fmt.Sprintf("Vous avez accès aux groupes : %s\nLe premier groupe sera utilisé.", strings.Join(userGroups, ", ")), fm.window)
	}

	// Rafraîchir la liste initiale
	fm.refreshFiles()
}

// Sélection / désélection de fichier
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

// Rafraîchir la liste des fichiers depuis le serveur
func (fm *FileManager) refreshFiles() {
	if fm.group == "" {
		dialog.ShowInformation("Erreur", "Aucun groupe sélectionné", fm.window)
		return
	}

	fm.progressBar.Show()
	fm.statusLabel.SetText("Chargement des fichiers...")

	go func() {
		newFileList := network.ListFiles(fm.group)
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

// Télécharger un fichier depuis le serveur avec ProgressBar
func (fm *FileManager) downloadFile() {
	if fm.selectedFile == "" {
		dialog.ShowInformation("Erreur", "Veuillez sélectionner un fichier", fm.window)
		return
	}

	dialog.ShowConfirm("Téléchargement", "Télécharger "+fm.selectedFile+" ?", func(confirmed bool) {
		if confirmed {
			fm.progressBar.SetValue(0)
			fm.progressBar.Show()
			fm.statusLabel.SetText("Téléchargement en cours...")

			go func() {
				err := network.DownloadWithProgress(fm.group, fm.selectedFile, func(bytesRead int64) {
					// Ici tu peux récupérer la taille totale côté serveur si besoin
					// Mais pour simplifier, on met juste 50% fixe si taille inconnue
					// Si taille connue, tu peux diviser par size totale
				})
				fm.progressBar.Hide()

				if err != nil {
					fm.statusLabel.SetText("Erreur lors du téléchargement")
					dialog.ShowError(err, fm.window)
					return
				}

				fm.statusLabel.SetText("Téléchargement terminé")
				dialog.ShowInformation("Succès", fm.selectedFile+" téléchargé avec succès", fm.window)
			}()
		}
	}, fm.window)
}

// Téléverser un fichier vers le serveur avec ProgressBar
func (fm *FileManager) uploadFile() {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil || reader == nil {
			return
		}

		fm.progressBar.SetValue(0)
		fm.progressBar.Show()
		fm.statusLabel.SetText("Téléversement en cours...")

		go func() {
			// Récupérer la taille du fichier
			fileStat, _ := os.Stat(reader.URI().Path())
			size := fileStat.Size()

			err := network.UploadWithProgress(fm.group, reader, func(bytesSent int64) {
				fm.progressBar.SetValue(float64(bytesSent) / float64(size))
			})

			fm.progressBar.Hide()
			if err != nil {
				fm.statusLabel.SetText("Erreur lors du téléversement")
				dialog.ShowError(err, fm.window)
				return
			}

			fm.statusLabel.SetText("Téléversement terminé")
			dialog.ShowInformation("Succès", reader.URI().Name()+" téléversé avec succès", fm.window)
		}()
	}, fm.window)
}

// Afficher l'historique serveur filtré par l'utilisateur

func (fm *FileManager) showServerHistory() {
	lines, err := network.GetServerHistory()
	if err != nil || len(lines) == 0 {
		dialog.ShowInformation("Historique serveur", "Aucun log disponible", fm.window)
		return
	}

	// Tableau de CanvasObjects
	objects := make([]fyne.CanvasObject, 0, len(lines)+1)

	// En-tête
	header := widget.NewLabel(fmt.Sprintf("%-12s %-8s %-10s %-30s %-10s", "Date", "Heure", "Opération", "Fichier", "Statut"))
	header.TextStyle = fyne.TextStyle{Bold: true}
	objects = append(objects, header)
	objects = append(objects, widget.NewSeparator())

	// Lignes du log
	for _, line := range lines {
		// Chaque ligne est au format CSV : Date,Heure,User,Group,Operation,Filename,Status
		parts := strings.Split(line, ",")
		if len(parts) < 7 {
			continue
		}
		label := widget.NewLabel(fmt.Sprintf("%-12s %-8s %-10s %-30s %-10s",
			parts[0],        // Date
			parts[1],        // Heure
			parts[4],        // Operation (UPLOAD/DOWNLOAD)
			parts[5],        // Filename
			parts[6],        // Status
		))
		objects = append(objects, label)
	}

	scroll := container.NewVScroll(container.NewVBox(objects...))
	scroll.SetMinSize(fyne.NewSize(1000, 600)) // Taille fenêtre agrandie

	dialog.ShowCustom("Historique serveur", "Fermer", scroll, fm.window)
}


// Affiche la fenêtre
func ShowFiles(a fyne.App) {
	fileManager := NewFileManager(a)
	fileManager.setupUI()
	fileManager.window.Show()
}
