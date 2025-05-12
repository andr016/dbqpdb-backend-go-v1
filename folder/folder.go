package folder

import (
	"fmt"
	"os"
)

func CreateFolderIfNotExist(folderPath string) error {
	// Check if the folder exists
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		// Create the folder if it doesn't exist
		err := os.Mkdir(folderPath, 0755) // 0755 is the permission
		if err != nil {
			return fmt.Errorf("failed to create folder: %v", err)
		}
		fmt.Println("Folder created:", folderPath)
	} else {
		fmt.Println("Folder already exists:", folderPath)
	}
	return nil
}
