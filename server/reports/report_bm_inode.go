package reports

import (
	structures "server/structures"
	utils "server/utils"
	"fmt"
	"os"
	"strings"
)

func ReportBMInode(superblock *structures.SuperBlock, diskPath string, path string) error {
	err:= utils.CreateParentDirs(path)
	if err != nil {
		return err
	}
	file, err := os.Open(diskPath)
	if err != nil {
		return err
	}
	defer file.Close()

	totalInodes := superblock.S_inodes_count + superblock.S_free_inodes_count

	var bitmapContent strings.Builder

	for i := int32(0); i < totalInodes; i++ {
		_, err := file.Seek(int64(superblock.S_bm_inode_start+i), 0)
		if err != nil {
			return fmt.Errorf("error al establecer el puntero en el archivo: %v", err)
		}

		char := make([]byte, 1)
		_, err = file.Read(char)
		if err != nil {
			return fmt.Errorf("error al leer el byte del archivo: %v", err)
		}

		bitmapContent.WriteByte(char[0])

		if (i+1)%20 == 0 {
			bitmapContent.WriteString("\n")	
		}
	}

	txtFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("error al crear el archivo TXT: %v", err)
	}
	defer txtFile.Close()

	_, err = txtFile.WriteString(bitmapContent.String())
	if err != nil {
		return fmt.Errorf("error al escribir en el archivo TXT: %v", err)
	}

	return nil;
}