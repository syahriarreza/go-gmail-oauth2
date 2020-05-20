package helper

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var rootDir = "D:/GO/src/eaciit/go-gmail-oauth2"
var ignoredFolder = "output"
var fileExt = ".js"
var isLogEnabled = true

// PrintPipe prints or save pipe into a text file
// ex: rezahelper.PrintPipe(pipes, "UpdatedHeadcount")
func PrintPipe(pipes interface{}, collection string, savedToFile bool) {
	jsonByte, err := json.MarshalIndent(pipes, "", "    ")
	if err != nil {
		log.Println("ERROR!", err)
	}
	if isLogEnabled {
		log.Println("PRINT Pipe\n", "db.getCollection('"+collection+"').aggregate(\n"+string(jsonByte)+"\n)")
	}
	if savedToFile {
		saveToFile(getCallerFnName(), collection, "db.getCollection('"+collection+"').aggregate(\n"+string(jsonByte)+"\n)")
	}
}

//PrintJSON PrintJSON
func PrintJSON(interF interface{}, message string, savedToFile bool) {
	jsonByte, _ := json.MarshalIndent(interF, "", "    ")
	if isLogEnabled {
		log.Println("PRINT JSON\n", string(jsonByte))
	}
	if savedToFile {
		saveToFile(getCallerFnName(), message, string(jsonByte))
	}
}

//TimeTrack TimeTrack
func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %ds", name, int64(elapsed/time.Second))
}

//TimeTrackMoreThanZeroSecond TimeTrackMoreThanZeroSecond
func TimeTrackMoreThanZeroSecond(start time.Time, name string) {
	elapsed := time.Since(start)
	if el := int64(elapsed / time.Second); el > 0 {
		log.Printf("%s took %ds", name, el)
	}
}

//TimeTrackFunc TimeTrackFunc
func TimeTrackFunc(start time.Time, message string) {
	elapsed := time.Since(start)
	if message == "" {
		log.Printf("EXEC-TIME: %s() took %.2fs", getCallerFnName(), float32(int64(elapsed/time.Millisecond))/1000)
	} else {
		log.Printf("EXEC-TIME: %s() %s took %.2fs", getCallerFnName(), message, float32(int64(elapsed/time.Millisecond))/1000)
	}
}

func printPipe(pipes interface{}, collection string) {
	jsonByte, err := json.MarshalIndent(pipes, "", "    ")
	if err != nil {
		log.Println("ERROR!", err)
	}
	log.Println("PRINT Pipe\n", "db.getCollection('"+collection+"').aggregate(\n"+string(jsonByte)+"\n)")
}

func getCallerFnName() string {
	// Skip this function, and fetch the PC and file for its parent
	pc, _, _, _ := runtime.Caller(2) //-- 0: this function -> GetCallerFnName()
	// Retrieve a Function object this functions parent
	functionObject := runtime.FuncForPC(pc)
	// Regex to extract just the function name (and not the module path)
	extractedFnName := regexp.MustCompile(`^.*\.(.*)$`)
	fnName := extractedFnName.ReplaceAllString(functionObject.Name(), "$1")
	return fnName
}

func generateFileName(fnName, message string) string {
	folderPath := filepath.Join(rootDir, ignoredFolder, "files")
	dir, err := os.Open(folderPath)
	if err != nil {
		log.Println("ERROR\n", err.Error())
	}

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		log.Println("ERROR\n", err.Error())
	}

	//--Prefix generation
	prefix := ""
	counter := 1
	for {
		fileIsFound := false
		prefix = fnName + "_" + strconv.Itoa(counter)
		for _, fileInfo := range fileInfos { //--loop each files
			if strings.Contains(fileInfo.Name(), prefix) {
				fileIsFound = true
			}
		}
		if !fileIsFound { //--if file is not found then use the prefix
			break
		}
		counter++
	}

	return prefix + " " + message + fileExt
}

func saveToFile(funcCallerName, message, str string) {
	fileName := generateFileName(funcCallerName, message)
	filePath := filepath.Join(rootDir, ignoredFolder, "files", fileName)
	f, err := os.Create(filePath)
	if err != nil {
		log.Println("ERROR Creating:", filePath, "\n", err.Error())
	}
	defer f.Close()
	f.Write([]byte(str))
}
