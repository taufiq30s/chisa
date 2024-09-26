package utils

import (
	"log"
	"os"
)

var (
	WarningLog *log.Logger
	InfoLog    *log.Logger
	ErrorLog   *log.Logger

	infoFile    *os.File
	warningFile *os.File
	errorFile   *os.File

	err error
)

func loadFile() {
	infoFile, err = os.OpenFile("logs/info.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	warningFile, err = os.OpenFile("logs/warning.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		log.Fatal(err)
	}
	errorFile, err = os.OpenFile("logs/error.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	loadFile()
	InfoLog = log.New(infoFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLog = log.New(warningFile, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(errorFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}
