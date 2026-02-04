package auth

import (
	"encoding/json"
	"errors"
	"os"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// fonction relative aux gro
func AddUser(username, password string) error {
	users, err := LoadUsers()
	if err != nil {
		return err
	}

	for _, u := range users {
		if u.Username == username {
			return errors.New("utilisateur déjà existant")
		}
	}

	users = append(users, User{
		Username: username,
		Password: password,
	})

	data, _ := json.MarshalIndent(users, "", "  ")
	return os.WriteFile("storage/users.json", data, 0644)
}

// fin
func LoadUsers() ([]User, error) {
	data, err := os.ReadFile("storage/users.json")
	if err != nil {
		return nil, err
	}

	var users []User
	err = json.Unmarshal(data, &users)
	return users, err
}

func Authenticate(username, password string) error {
	users, err := LoadUsers()
	if err != nil {
		return err
	}

	for _, user := range users {
		if user.Username == username && user.Password == password {
			return nil
		}
	}
	return errors.New("authentification échouée")
}
