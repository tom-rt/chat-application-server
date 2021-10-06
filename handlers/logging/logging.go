package logging

import (
	"log"
	"os"
	"time"
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
	t := time.Now()
	now := t.String()
	completeLog := now + " : " + toLog

	_, err := logFile.Write([]byte(completeLog))
	if err != nil {
		log.Println("Error writing to log file:", err)
	}
}
