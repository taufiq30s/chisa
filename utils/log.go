package utils

import (
	"log"
	"os"
)

var (
	WarningLog *log.Logger
	InfoLog    *log.Logger
	ErrorLog   *log.Logger
	DebugLog   *log.Logger

	infoFile    *os.File
	warningFile *os.File
	errorFile   *os.File
	debugFile   *os.File

	err error
)

func loadFile() {
	infoFile, err = os.OpenFile("logs/info.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, FilePermissionReadWrite)
	if err != nil {
		log.Fatal(err)
	}
	warningFile, err = os.OpenFile("logs/warning.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, FilePermissionReadWriteGroup)
	if err != nil {
		log.Fatal(err)
	}
	errorFile, err = os.OpenFile("logs/error.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, FilePermissionReadWriteGroup)
	if err != nil {
		log.Fatal(err)
	}
	debugFile, err = os.OpenFile("logs/debug.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, FilePermissionReadWrite)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	loadFile()
	InfoLog = log.New(infoFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLog = log.New(warningFile, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(errorFile, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	DebugLog = log.New(debugFile, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)
}
