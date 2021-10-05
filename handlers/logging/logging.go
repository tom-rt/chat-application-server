package logging

import (
	"log"
	"os"
)

var logFile *os.File

func InitLog() error {
	var err error
	logFile, err = os.OpenFile("history.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error opening file:", err)
		return err
	}
	return nil
}

func WriteLog(toLog string) {
	_, err := logFile.Write([]byte(toLog))
	if err != nil {
		log.Println("Error writing to log file:", err)
	}
}
