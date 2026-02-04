package storage

import (
	"encoding/json"
	"file_transfert_go/serveur/authentification"
	
	"os"
)
func GetAllUsers() []string {
	users, err := auth.LoadUsers()
	if err != nil {
		return []string{}
	}

	var list []string
	for _, u := range users {
		list = append(list, u.Username)
	}
	return list
}




func DeleteUser(username string) error {
	users, err := LoadUsers()
	if err != nil {
		return err
	}

	delete(users, username)
	return SaveUsers(users)
}
const usersFile = "storage/users.json"

// LoadUsers lit tous les utilisateurs depuis users.json
func LoadUsers() (map[string]string, error) {
	users := make(map[string]string)

	data, err := os.ReadFile(usersFile)
	if err != nil {
		if os.IsNotExist(err) {
			return users, nil // fichier inexistant -> map vide
		}
		return nil, err
	}

	err = json.Unmarshal(data, &users)
	return users, err
}

// SaveUsers sauvegarde tous les utilisateurs dans users.json
func SaveUsers(users map[string]string) error {
	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	// créer dossier si nécessaire
	os.MkdirAll("storage", 0755)

	return os.WriteFile(usersFile, data, 0644)
}
