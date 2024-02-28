package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const totalFile = 100

var tempPath = filepath.Join(os.Getenv("TEMP"), "test-folder")

type NameInfo struct {
	Name      string
	IsChanged bool
}

func main() {
	log.Println("start")
	start := time.Now()

	createFiles()

	chanNames := readFiles()

	chanRename1 := rename(chanNames)
	chanRename2 := rename(chanNames)
	chanRename3 := rename(chanNames)
	chanRename4 := rename(chanNames)
	chanRename := mergeChanFileInfo(chanRename1, chanRename2, chanRename3, chanRename4)

	counterRenamed := 0
	counterTotal := 0
	for fileInfo := range chanRename {
		if fileInfo.IsChanged {
			counterRenamed++
		}
		counterTotal++
	}

	log.Printf("%d/%d files renamed", counterRenamed, counterTotal)

	duration := time.Since(start)
	log.Println("done in", duration.Seconds(), "seconds")
}

func createFiles() {
	os.RemoveAll(tempPath)
	os.MkdirAll(tempPath, os.ModePerm)

	for i := 0; i < totalFile; i++ {
		filename := filepath.Join(tempPath, fmt.Sprintf("name-%d.txt", i))
		content := "test"
		err := os.WriteFile(filename, []byte(content), os.ModePerm)
		if err != nil {
			log.Println("Error writing file", filename)
		}

		if i%100 == 0 && i > 0 {
			log.Println(i, "files created", content)
		}
	}

	log.Printf("%d of total files created", totalFile)
}

func readFiles() <-chan NameInfo {
	chanOut := make(chan NameInfo)

	go func() {
		err := filepath.Walk(tempPath, func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			chanOut <- NameInfo{
				Name: path,
			}

			return nil
		})
		if err != nil {
			log.Println("ERROR:", err.Error())
		}

		close(chanOut)
	}()

	return chanOut
}

func rename(chanIn <-chan NameInfo) <-chan NameInfo {
	chanOut := make(chan NameInfo)

	go func() {
		for infoName := range chanIn {
			newPath := filepath.Join(tempPath, fmt.Sprintf("test-%v.txt", time.Now().UnixMicro()))
			err := os.Rename(infoName.Name, newPath)
			infoName.IsChanged = err == nil
			chanOut <- infoName
		}

		close(chanOut)
	}()

	return chanOut
}

func mergeChanFileInfo(chanInMany ...<-chan NameInfo) <-chan NameInfo {
	wg := new(sync.WaitGroup)
	chanOut := make(chan NameInfo)

	wg.Add(len(chanInMany))
	for _, eachChan := range chanInMany {
		go func(eachChan <-chan NameInfo) {
			for eachChanData := range eachChan {
				chanOut <- eachChanData
			}
			wg.Done()
		}(eachChan)
	}

	go func() {
		wg.Wait()
		close(chanOut)
	}()

	return chanOut
}
