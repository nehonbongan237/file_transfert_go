package storage

import (
	"fmt"
	"os"
	"time"
)

// Ajouter un log
func LogOperation(user, group, filename, operation, status string) error {
	os.MkdirAll("logs", 0755)
	f, err := os.OpenFile("logs/operations.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	// Format CSV simple : Date,Heure,User,Group,Operation,Filename,Status
	now := time.Now()
	line := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s\n",
		now.Format("2006-01-02"),
		now.Format("15:04:05"),
		user, group, operation, filename, status)

	_, err = f.WriteString(line)
	return err
}
