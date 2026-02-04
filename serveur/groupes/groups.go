package groups

import (
	"encoding/json"
	"errors"
	auth "file_transfert_go/serveur/authentification"
	"os"
)

type Group struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
}

func LoadGroups() ([]Group, error) {
	data, err := os.ReadFile("storage/groups.json")
	if err != nil {
		return nil, err
	}

	var groups []Group
	err = json.Unmarshal(data, &groups)
	return groups, err
}

func UserInGroup(username, groupName string) bool {
	groups, err := LoadGroups()
	if err != nil {
		return false
	}

	for _, g := range groups {
		if g.Name == groupName {
			for _, m := range g.Members {
				if m == username {
					return true
				}
			}
		}
	}
	return false
}

func GetUserGroups(username string) []string {
	groups, _ := LoadGroups()
	var result []string

	for _, g := range groups {
		for _, m := range g.Members {
			if m == username {
				result = append(result, g.Name)
			}
		}
	}
	return result
}

// AddUser ajoute un utilisateur (par nom) au groupe choisi et persiste dans storage/groups.json.
func AddUser(username, groupName string) error {
	groups, err := LoadGroups()
	if err != nil {
		return err
	}

	for _, g := range groups {
		for _, m := range g.Members {
			if m == username {
				return errors.New("utilisateur déjà présent dans un groupe")
			}
		}
	}

	added := false
	for i := range groups {
		if groups[i].Name == groupName {
			groups[i].Members = append(groups[i].Members, username)
			added = true
			break
		}
	}

	if !added {
		groups = append(groups, Group{
			Name:    groupName,
			Members: []string{username},
		})
	}

	data, _ := json.MarshalIndent(groups, "", "  ")
	return os.WriteFile("storage/groups.json", data, 0644)
}

// AddUserWithCredentials ajoute l'utilisateur dans users.json et l'associe au groupe choisi.
func AddUserWithCredentials(username, password, groupName string) error {
	if err := auth.AddUser(username, password); err != nil {
		return err
	}

	return AddUser(username, groupName)
}
