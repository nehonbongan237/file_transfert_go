package network

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"fyne.io/fyne/v2" 
	
)

//nst ServerAddr = "localhost:8080"

var Conn net.Conn
var Reader *bufio.Reader
var Token string
func Connect() error { 
		serverAddr, err := getAddress()
		if err != nil {
			return err
		}
		var er error
		Conn, er = net.Dial("tcp", serverAddr)

	 if er != nil { return err }
	  Reader = bufio.NewReader(Conn)
	   return nil
	 }

 func getAddress() (string, error) {
	addr := net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 9999,
	}

	conn, err := net.ListenUDP("udp4", &addr)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// ⏱️ timeout pour éviter le blocage
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	buf := make([]byte, 1024)
	n, remoteAddr, err := conn.ReadFromUDP(buf)
	if err != nil {
		return "", fmt.Errorf("serveur introuvable (timeout)")
	}

	msg := strings.TrimSpace(string(buf[:n]))

	if !strings.HasPrefix(msg, "SERVER:") {
		return "", fmt.Errorf("message UDP invalide")
	}

	port := strings.TrimPrefix(msg, "SERVER:")
	serverIP := remoteAddr.IP.String()

	return fmt.Sprintf("%s:%s", serverIP, port), nil
}


 
func Authenticate(user, pass string) string {
	fmt.Fprintf(Conn, "AUTH %s %s\n", user, pass)
	resp, _ := Reader.ReadString('\n')

	if strings.HasPrefix(resp, "AUTH_OK") {
		Token = strings.TrimSpace(strings.TrimPrefix(resp, "AUTH_OK "))
		return Token
	}
	return ""
}

func ListFiles(group string) []string {
	fmt.Fprintf(Conn, "LIST %s %s\n", Token, group)
	resp, _ := Reader.ReadString('\n')

	if !strings.HasPrefix(resp, "OK") {
		return nil
	}

	files := strings.TrimSpace(strings.TrimPrefix(resp, "OK "))
	if files == "" {
		return nil
	}
	return strings.Split(files, ",")
}

func Download(group, filename string) error {
	fmt.Fprintf(Conn, "GET %s %s %s\n", Token, group, filename)

	// Lire l'en-tête texte (taille du fichier)
	headerReader := bufio.NewReader(Conn)
	header, err := headerReader.ReadString('\n')
	if err != nil {
		return err
	}

	if !strings.HasPrefix(header, "OK") {
		return fmt.Errorf("erreur serveur: %s", header)
	}

	sizeStr := strings.TrimSpace(strings.TrimPrefix(header, "OK "))
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return fmt.Errorf("taille invalide: %v", err)
	}

	os.MkdirAll("downloads", 0755)
	path := filepath.Join("downloads", filename)

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copier exactement "size" octets depuis Conn vers le fichier
	// On utilise Conn directement, pas headerReader, pour éviter de mélanger texte et binaire
	_, err = io.CopyN(file, Conn, size)
	if err != nil {
		return err
	}

	return nil
}

func Upload(group string, file fyne.URIReadCloser) error {
	// Ouvrir le fichier local
	f, err := os.Open(file.URI().Path())
	if err != nil {
		return err
	}
	defer f.Close()

	// Taille et nom du fichier
	stat, err := f.Stat()
	if err != nil {
		return err
	}
	size := stat.Size()
	filename := filepath.Base(f.Name())

	// Envoyer la commande UPLOAD avec token, groupe, nom et taille
	fmt.Fprintf(Conn, "UPLOAD %s %s %s %d\n", Token, group, filename, size)

	// Attendre que le serveur soit prêt
	resp, err := Reader.ReadString('\n')
	if err != nil {
		return err
	}
	resp = strings.TrimSpace(resp)
	if resp != "READY" {
		return fmt.Errorf("serveur non prêt pour upload : %s", resp)
	}

	// Envoyer le contenu du fichier par buffer
	buf := make([]byte, 4096)
	for {
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}
		_, err = Conn.Write(buf[:n])
		if err != nil {
			return err
		}
	}

	// Lire la confirmation finale
	finalResp, err := Reader.ReadString('\n')
	if err != nil {
		return err
	}
	finalResp = strings.TrimSpace(finalResp)
	if finalResp != "OK UPLOAD_SUCCESS" {
		return fmt.Errorf("upload échoué : %s", finalResp)
	}

	return nil
}
func GetUserGroups() ([]string, error) {
	fmt.Fprintf(Conn, "MYGROUPS %s\n", Token)
	resp, err := Reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	resp = strings.TrimSpace(resp)
	if !strings.HasPrefix(resp, "OK") {
		return nil, fmt.Errorf("erreur serveur: %s", resp)
	}

	groupsStr := strings.TrimPrefix(resp, "OK ")
	if groupsStr == "" {
		return []string{}, nil
	}
	return strings.Split(groupsStr, ","), nil
}
func GetServerHistory() ([]string, error) {
	fmt.Fprintf(Conn, "HISTORY %s\n", Token)
	resp, err := Reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	resp = strings.TrimSpace(resp)
	if !strings.HasPrefix(resp, "OK") {
		return nil, fmt.Errorf("erreur serveur: %s", resp)
	}

	historyStr := strings.TrimPrefix(resp, "OK ")
	if historyStr == "" {
		return []string{}, nil
	}

	// Séparer les lignes
	lines := strings.Split(historyStr, "|")
	return lines, nil
}
// DownloadWithProgress télécharge un fichier depuis le serveur et appelle progress(bytesRead)
func DownloadWithProgress(group, filename string, progress func(int64)) error {
	fmt.Fprintf(Conn, "GET %s %s %s\n", Token, group, filename)

	// Lire l'en-tête (taille)
	headerReader := bufio.NewReader(Conn)
	header, err := headerReader.ReadString('\n')
	if err != nil {
		return err
	}

	if !strings.HasPrefix(header, "OK") {
		return fmt.Errorf("erreur serveur: %s", header)
	}

	sizeStr := strings.TrimSpace(strings.TrimPrefix(header, "OK "))
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
		return fmt.Errorf("taille invalide: %v", err)
	}

	os.MkdirAll("downloads", 0755)
	path := filepath.Join("downloads", filename)

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Lecture par buffer et mise à jour du callback
	buf := make([]byte, 4096)
	var totalRead int64 = 0
	for totalRead < size {
		n, err := Conn.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		_, err = file.Write(buf[:n])
		if err != nil {
			return err
		}

		totalRead += int64(n)
		progress(totalRead) // Met à jour la ProgressBar
	}

	return nil
}

// UploadWithProgress téléverse un fichier vers le serveur et appelle progress(bytesSent)
func UploadWithProgress(group string, file fyne.URIReadCloser, progress func(int64)) error {
	f, err := os.Open(file.URI().Path())
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}
	size := stat.Size()
	filename := filepath.Base(f.Name())

	// Envoyer la commande UPLOAD
	fmt.Fprintf(Conn, "UPLOAD %s %s %s %d\n", Token, group, filename, size)

	resp, err := Reader.ReadString('\n')
	if err != nil {
		return err
	}
	resp = strings.TrimSpace(resp)
	if resp != "READY" {
		return fmt.Errorf("serveur non prêt pour upload : %s", resp)
	}

	// Envoyer par buffer avec callback
	buf := make([]byte, 4096)
	var totalSent int64 = 0
	for {
		n, err := f.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		_, err = Conn.Write(buf[:n])
		if err != nil {
			return err
		}

		totalSent += int64(n)
		progress(totalSent) // Met à jour la ProgressBar
	}

	// Confirmation finale
	finalResp, err := Reader.ReadString('\n')
	if err != nil {
		return err
	}
	finalResp = strings.TrimSpace(finalResp)
	if finalResp != "OK UPLOAD_SUCCESS" {
		return fmt.Errorf("upload échoué : %s", finalResp)
	}

	return nil
}
