package stores

import (
	"fmt"
	"server/console"
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

// Función para debug - imprimir discos cargados
func PrintLoadedDisks() {
	console.PrintInfo("📀 Discos cargados actualmente:")
	if len(LoadedDiskPaths) == 0 {
		console.PrintWarning("  ⚠️ No hay discos cargados")
		return
	}

	for letter, path := range LoadedDiskPaths {
		console.PrintInfo(fmt.Sprintf("  %s -> %s", letter, path))
	}
}

// Función para debug - imprimir particiones montadas
func PrintMountedPartitions() {
	console.PrintInfo("🗂️ Particiones montadas actualmente:")
	if len(MountedPartitions) == 0 {
		console.PrintWarning("  ⚠️ No hay particiones montadas")
		return
	}

	for id, path := range MountedPartitions {
		console.PrintInfo(fmt.Sprintf("  %s -> %s", id, path))
	}
}
