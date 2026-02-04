package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
)

var sessions = make(map[string]string)
var mutex sync.Mutex

// Génère un token aléatoire
func generateToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Crée une session
func CreateSession(username string)string {
	mutex.Lock()
	defer mutex.Unlock()

	token := generateToken()
	sessions[token] = username
	return token
}

// Vérifie un token
func GetUserByToken(token string) (string, bool) {
	mutex.Lock()
	defer mutex.Unlock()
	user, ok := sessions[token]
	return user, ok
}
