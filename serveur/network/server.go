package network

import (
	"bufio"
	auth "file_transfert_go/serveur/authentification"
	groups "file_transfert_go/serveur/groupes"
	"file_transfert_go/serveur/storage"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
)

func StartServer(address string) {
	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		log.Fatal("Erreur serveur:", err)
	}
	defer listener.Close()

	log.Println("Serveur en écoute sur", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Erreur connexion:", err)
			continue
		}

		go handleClient(conn)

	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	log.Println("Nouvelle connexion de", conn.RemoteAddr().String())
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "AUTH":
			if len(parts) != 3 {
				conn.Write([]byte("ERROR Format AUTH\n"))
				continue
			}

			err := auth.Authenticate(parts[1], parts[2])
			if err != nil {
				conn.Write([]byte("AUTH_FAIL\n"))
			} else {
				//modifier la commande AUTH pour qu"elle renvoi un token de session
				token := auth.CreateSession(parts[1])
				conn.Write([]byte("AUTH_OK " + token + "\n"))
			}

		case "LIST":
			if len(parts) != 3 {
				conn.Write([]byte("ERROR LIST_FORMAT\n"))
				continue
			}

			token := parts[1]
			group := parts[2]

			user, ok := auth.GetUserByToken(token)
			if !ok {
				conn.Write([]byte("ERROR INVALID_TOKEN\n"))
				continue
			}

			if !groups.UserInGroup(user, group) {
				conn.Write([]byte("ERROR ACCESS_DENIED\n"))
				continue
			}

			files, err := storage.ListFiles(group)
			if err != nil {
				conn.Write([]byte("ERROR NO_GROUP_DIR\n"))
				continue
			}

			conn.Write([]byte("OK " + strings.Join(files, ",") + "\n"))

		case "MYGROUPS":
			if len(parts) != 2 {
				conn.Write([]byte("ERROR MYGROUPS_FORMAT\n"))
				continue
			}

			user, ok := auth.GetUserByToken(parts[1])
			if !ok {
				conn.Write([]byte("ERROR INVALID_TOKEN\n"))
				continue
			}

			userGroups := groups.GetUserGroups(user)
			conn.Write([]byte("OK " + strings.Join(userGroups, ",") + "\n"))
		case "UPLOAD":
	if len(parts) < 5 {
		conn.Write([]byte("ERROR UPLOAD_FORMAT\n"))
		continue
	}

	token := parts[1]
	group := parts[2]
	sizeStr := parts[len(parts)-1]                       // dernier élément = taille
	filename := strings.Join(parts[3:len(parts)-1], " ") // tout ce qui est entre group et size = nom

	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		conn.Write([]byte("ERROR INVALID_SIZE\n"))
		continue
	}

	user, ok := auth.GetUserByToken(token)
	if !ok {
		conn.Write([]byte("ERROR INVALID_TOKEN\n"))
		continue
	}

	if !groups.UserInGroup(user, group) {
		conn.Write([]byte("ERROR ACCESS_DENIED\n"))
		continue
	}

	conn.Write([]byte("READY\n"))

	limited := io.LimitReader(conn, int64(size))
	err = storage.SaveFile(group, filename, limited)
	if err != nil {
		conn.Write([]byte("ERROR UPLOAD_FAILED\n"))
						storage.LogOperation(user, group, filename, "UPLOAD","FAILED")
		continue
	}

	conn.Write([]byte("OK UPLOAD_SUCCESS\n"))
	storage.LogOperation(user, group, filename, "UPLOAD", "SUCCESS")


		case "GET":
			if len(parts) < 4 {
				conn.Write([]byte("ERROR GET_FORMAT\n"))
				continue
			}

			token := parts[1]
			group := parts[2]
			filename := strings.Join(parts[3:], " ") // Tout le reste = filename


			user, ok := auth.GetUserByToken(token)
			if !ok {
				conn.Write([]byte("ERROR INVALID_TOKEN\n"))
				continue
			}

			if !groups.UserInGroup(user, group) {
				conn.Write([]byte("ERROR ACCESS_DENIED\n"))
				continue
			}

			file, size, err := storage.OpenFile(group, filename)
			if err != nil {
				conn.Write([]byte("ERROR FILE_NOT_FOUND\n"))
				storage.LogOperation(user, group, filename, "DOWNLOAD", "FAILED")

				continue
			}
			defer file.Close()

			conn.Write([]byte("OK " + strconv.FormatInt(size, 10) + "\n"))

			io.Copy(conn, file)
			storage.LogOperation(user, group, filename, "DOWNLOAD", "SUCCESS")


		default:
			conn.Write([]byte("UNKNOWN_COMMAND\n"))
		}
	}
}
