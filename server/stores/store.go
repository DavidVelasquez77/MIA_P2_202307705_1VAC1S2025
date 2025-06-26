package stores

import (
	"errors"
	"fmt"
	"path/filepath"
	"server/console"
	"server/structures"
	"server/utils"
	"strings"
)

const Carnet string = "05"                                                           //2023007705
const PathDisk string = "/home/vela/Documentos/MIA/MIA_P2_202307705_1VAC1S2025/test" //FIXME cambiar el path

var (
	MountedPartitions map[string]string = make(map[string]string) //ID:path
	LogedIdPartition  string            = ""
	LogedUser         string            = ""
	LoadedDiskPaths   map[string]string = make(map[string]string) //Nombre:path
)

func GetPathDisk(name string) string {
	return fmt.Sprintf(`%s/%s.dsk`, PathDisk, name)
}

func GetMountedPartition(id string) (*structures.PARTITION, string, error) {
	path := MountedPartitions[id]
	if path == "" {
		return nil, "", errors.New("la particion no esta montada")
	}
	var mbr structures.MBR

	err := mbr.DeserializeMBR(path)
	if err != nil {
		return nil, "", err
	}

	partition, _, err := mbr.GetPartitionByID(id)
	if partition == nil {
		return nil, "", err
	}
	return partition, path, nil

}

func DeleteMountedPartitions(path string) {
	for key, value := range MountedPartitions {
		if value == path {
			delete(MountedPartitions, key)
			delete(utils.PathToPartitionCount, path)
			// delete(utils.PathToLetter, path)
		}
	}
	for key, value := range LoadedDiskPaths {
		if value == path {
			delete(LoadedDiskPaths, key)
		}
	}
}

// Nueva función para limpiar solo los discos cargados
func RemoveLoadedDisk(path string) {
	for key, value := range LoadedDiskPaths {
		if value == path {
			delete(LoadedDiskPaths, key)
		}
	}
}

func GetMountedPartitionRep(id string) (*structures.MBR, *structures.SuperBlock, string, error) {
	path := MountedPartitions[id]
	if path == "" {
		return nil, nil, "", errors.New("la particion no esta montada")
	}

	var mbr structures.MBR
	err := mbr.DeserializeMBR(path)
	if err != nil {
		return nil, nil, "", err
	}
	partition, _, err := mbr.GetPartitionByID(id)
	if partition == nil {
		return nil, nil, "", err
	}

	var sb structures.SuperBlock

	err = sb.Deserialize(path, int64(partition.Part_start))
	if err != nil {
		return nil, nil, "", err
	}

	return &mbr, &sb, path, nil

}

func GetNameDisk(idDisk string) string {
	pathDisk := MountedPartitions[idDisk]
	baseName := strings.TrimSuffix(filepath.Base(pathDisk), filepath.Ext(pathDisk))
	return baseName
}

func GetMountedPartitionSuperblock(id string) (*structures.SuperBlock, *structures.PARTITION, string, error) {
	path := MountedPartitions[id]
	if path == "" {
		return nil, nil, "", errors.New("la particion no esta montada")
	}

	var mbr structures.MBR

	err := mbr.DeserializeMBR(path)
	if err != nil {
		return nil, nil, "", err
	}

	partition, _, err := mbr.GetPartitionByID(id)
	if err != nil {
		return nil, nil, "", err
	}

	var sb structures.SuperBlock

	err = sb.Deserialize(path, int64(partition.Part_start))
	if err != nil {
		return nil, nil, "", err
	}

	return &sb, partition, path, nil
}

// Nueva función para limpiar completamente el estado
func ClearAllDisks() {
	LoadedDiskPaths = make(map[string]string)
	MountedPartitions = make(map[string]string)
	LogedIdPartition = ""
	LogedUser = ""
}

// Función mejorada para debug del estado actual
func PrintCurrentState() {
	console.PrintInfo("=== ESTADO ACTUAL DEL SISTEMA ===")
	console.PrintInfo(fmt.Sprintf("📀 Discos cargados: %d", len(LoadedDiskPaths)))
	for letter, path := range LoadedDiskPaths {
		console.PrintInfo(fmt.Sprintf("  - %s: %s", letter, path))
	}

	console.PrintInfo(fmt.Sprintf("🗂️ Particiones montadas: %d", len(MountedPartitions)))
	for id, path := range MountedPartitions {
		console.PrintInfo(fmt.Sprintf("  - %s: %s", id, path))
	}

	console.PrintInfo(fmt.Sprintf("👤 Usuario logueado: %s", LogedUser))
	console.PrintInfo(fmt.Sprintf("💾 Partición logueada: %s", LogedIdPartition))
	console.PrintSeparator()
}
