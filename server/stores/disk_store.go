package stores

import (
	"fmt"
	"server/structures"
)

func GetAllLoadedDisks() map[string]string {
	return LoadedDiskPaths
}

func GetDiskInfo(diskLetter string) (*structures.MBR, string, error) {
	diskPath, exists := LoadedDiskPaths[diskLetter]
	if !exists {
		return nil, "", fmt.Errorf("disco %s no encontrado", diskLetter)
	}

	mbr := &structures.MBR{}
	err := mbr.DeserializeMBR(diskPath)
	if err != nil {
		return nil, "", err
	}

	return mbr, diskPath, nil
}
