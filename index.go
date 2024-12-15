package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func main() {
	// Ensure the terminal remains open after execution
	defer func() {
		fmt.Println("Press Enter to exit...")
		fmt.Scanln()
	}()

	// if no path is set (isn!t launched from context menu)
	if len(os.Args) < 2 {

		shellKeyPath := `*\shell\where_was_file_downloaded_from`
		commandKeyPath := shellKeyPath + `\command`

		// Get the current executable path
		programPath, err := os.Executable()
		if err != nil {
			log.Fatalf("Couldn't get current working directory: %v", err)
		}

		programPathCommand := programPath + " " + `"%1"` //C:/Idk/hentai/program.exe "%1"

		// Check if the `shell` key exists
		exists, err := keyExists(registry.CLASSES_ROOT, shellKeyPath)
		if err != nil {
			log.Fatalf("Failed to check if key exists: %v", err)
		}

		// If it doesn't exist, create the required registry keys
		if !exists {
			err := createKeys(registry.CLASSES_ROOT, shellKeyPath, commandKeyPath, programPathCommand)
			if err != nil {
				log.Println("You should probably launch this program with administrator! (Writing into registry requires it)")
                fmt.Scanln() //defer func() at the beginning doesn't work on errors :)
				log.Fatalf("Failed to create keys: %v", err)
			}
		}

		// Verify if the program path is correct
		isCorrectProgramPath, err := checkIfKeyValueIsCorrect(registry.CLASSES_ROOT, commandKeyPath, programPathCommand)
		if err != nil {
			log.Fatalf("Failed to check if key value is correct: %v", err)
		}

		// Update the key value if it's incorrect
		if !isCorrectProgramPath {
			err := fixKeyValue(registry.CLASSES_ROOT, commandKeyPath, programPathCommand)
			if err != nil {
				log.Println("You should probably launch this program with administrator! (Writing into registry requires it)")
                fmt.Scanln()
				log.Fatalf("Failed to fix key value: %v", err)
			}
		}

		log.Println("Registry wise, it's all set :3")
		log.Println("Use context menu to use this program!")
		log.Print(`(Right click a file > "Where was this file downloaded from?")`)
	}

	filePath := os.Args[1]

	log.Printf("Choosen path: %s\n\n", filePath)

    infoPath := filePath + ":Zone.Identifier"

	file, err := os.ReadFile(infoPath)
	if err != nil {
        if os.IsNotExist(err){
            log.Println("This file is probably not downloaded from the internet. :(")
            return
        }
		log.Fatalf("Failed to open the file: %v", err)
	}

	fileContents := string(file)
	//log.Println(fileContents)

	url := strings.Split(fileContents, "HostUrl=")[1]
	if strings.TrimSpace(url) == "about:internet" {
		log.Println("This file was probably downloaded in incognito mode :(")
		fmt.Println("(doesn't store the URL, at least I don't know about it)")
		return
	}

	log.Printf("Found URL: %s\n", url)
}

// keyExists checks if a registry key exists
func keyExists(baseKey registry.Key, subKeyPath string) (bool, error) {
	key, err := registry.OpenKey(baseKey, subKeyPath, registry.READ)
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, err
	}
	defer key.Close()

	return true, nil
}

// createKeys create the registry keys in "Počítač\HKEY_CLASSES_ROOT\*\shell\""
func createKeys(baseKey registry.Key, keyPath string, commandKeyPath string, programPath string) error {
	shellKey, _, err := registry.CreateKey(baseKey, keyPath, registry.WRITE)
	if err != nil {
		return err
	}
	defer shellKey.Close()

	if err := shellKey.SetStringValue("", "Where was this file downloaded from?"); err != nil {
		return err
	}

	commandKey, _, err := registry.CreateKey(baseKey, commandKeyPath, registry.WRITE)
	if err != nil {
		return err
	}
	defer commandKey.Close()

	if err := commandKey.SetStringValue("", programPath); err != nil {
		return err
	}

	return nil
}

// checkIfKeyValueIsCorrect checks if the program value is correctly set to this programs path in "Počítač\HKEY_CLASSES_ROOT\*\shell\shellKeyPath\commandKeyPath"
func checkIfKeyValueIsCorrect(baseKey registry.Key, commandKeyPath string, programPath string) (bool, error) {
	key, err := registry.OpenKey(baseKey, commandKeyPath, registry.READ)
	if err != nil {
		return false, err
	}
	defer key.Close()

	val, _, err := key.GetStringValue("")
	if err != nil {
		if err == registry.ErrNotExist {
			return false, nil
		}
		return false, err
	}

	return val == programPath, nil
}

// fixKeyValue corrects the registry key path to program to the current program path
func fixKeyValue(baseKey registry.Key, commandKeyPath string, programPath string) error {
	key, err := registry.OpenKey(baseKey, commandKeyPath, registry.WRITE)
	if err != nil {
		return err
	}
	defer key.Close()

	err = key.SetStringValue("", programPath)
	if err != nil {
		return err
	}

	return nil
}
