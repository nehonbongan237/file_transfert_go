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
	"fyne.io/fyne/v2" 
	"time"
)

const ServerAddr = "localhost:8080"

var Conn net.Conn
var Reader *bufio.Reader
var Token string

func Connect() error {
	var err error
	Conn, err = net.Dial("tcp", ServerAddr)
	if err != nil {
		return err
	}
	Reader = bufio.NewReader(Conn)
	return nil
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
func LogClientOperation(operation, group, filename, status string) {
	os.MkdirAll("client_logs", 0755)
	f, _ := os.OpenFile("client_logs/operations.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	now := time.Now()
	line := fmt.Sprintf("%s,%s,%s,%s,%s\n",
		now.Format("2006-01-02"),
		now.Format("15:04:05"),
		operation, group, filename, status)

	f.WriteString(line)
}
