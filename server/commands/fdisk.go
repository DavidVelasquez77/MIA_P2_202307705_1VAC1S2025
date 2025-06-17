package commands

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	stores "server/stores"
	"server/structures"
	"server/utils"
	"strconv"
	"strings"
)

type FDISK struct {
	size   int
	unit   string
	fit    string
	path   string
	typ    string
	name   string
	delete string
	add    int
}

func ParseFdisk(tokens []string) (string, error) {
	cmd := &FDISK{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-size=\d+|-unit=[kKmMbB]|-fit=[bBfF]{2}|-driveletter=[A-Za-z]|-type=[pPeElL]|-name="[^"]+"|-name=[^\s]+|-delete=[fFuUlL]+|-add=-?\d+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parametro invalid: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-size":
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return "", err
			}
			cmd.size = size
		case "-unit":
			value = strings.ToUpper(value)
			if value != "K" && value != "M" && value != "B" {
				return "", errors.New("la unidad debe ser K o M o B")
			}
			cmd.unit = strings.ToUpper(value)
		case "-fit":
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return "", errors.New("el ajuste debe ser BF, FF o WF")
			}
			cmd.fit = value
		case "-driveletter":
			if value == "" {
				return "", errors.New("el driveletter no puede estar vacío")
			}
			cmd.path = value
		case "-type":
			value = strings.ToUpper(value)
			if value != "P" && value != "E" {
				return "", errors.New("el tipo debe ser P o E")
			}
			cmd.typ = value
		case "-name":
			if value == "" {
				return "", errors.New("el nombre no puede estar vacío")
			}
			cmd.name = value
		case "-delete":
			value = strings.ToLower(value)
			if value != "fast" && value != "full" {
				return "", errors.New("para -delete se debe de indicar si sera fast o full")
			}
			cmd.delete = value
		case "-add":
			size, err := strconv.Atoi(value)
			if err != nil {
				return "", err
			}
			cmd.add = size
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.delete == "" && cmd.add == 0 {
		if cmd.size == 0 {
			return "", errors.New("faltan parámetros requeridos: -size")
		}
	}
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -driveletter")
	}
	cmd.path = stores.GetPathDisk(cmd.path)
	if cmd.name == "" {
		return "", errors.New("faltan parámetros requeridos: -name")
	}

	if cmd.unit == "" {
		cmd.unit = "K"
	}

	if cmd.fit == "" {
		cmd.fit = "WF"
	}

	if cmd.typ == "" {
		cmd.typ = "P"
	}
	if cmd.delete != "" && cmd.add != 0 {
		return "", errors.New("no se puede tener add y delete en el mismo comando")
	}

	if cmd.delete != "" {
		outcome := askConsent()
		if outcome {
			err := deletePartition(cmd)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("FDISK: %s delete exitosamente", cmd.name), nil
		}
		return fmt.Sprintf("FDISK: %s delete cancelada exitosamente", cmd.name), nil
	} else if cmd.add != 0 {
		err := addPartition(cmd)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("FDISK: %s add exitosamente", cmd.name), nil
	} else {
		err := commandFdisk(cmd)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("FDISK: %s creado exitosamente", cmd.name), nil
	}

}

func askConsent() bool {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Desea confirmar la ejecucion del delete? [y/n]: ")
		if !scanner.Scan() {
			break
		}
		input := scanner.Text()
		input = strings.ToLower(input)
		if input == "y" {
			return true
		} else if input == "n" {
			return false
		}
	}
	return false
}

func commandFdisk(fdisk *FDISK) error {
	sizeBytes, err := utils.ConvertToBytes(fdisk.size, fdisk.unit)
	if err != nil {
		return err
	}

	if fdisk.typ == "P" {
		err = createPrimaryPartition(fdisk, sizeBytes)
		if err != nil {
			return err
		}

	} else if fdisk.typ == "E" {
		err = createExtendedPartittion(fdisk, sizeBytes)
		if err != nil {
			return err
		}
	}
	return nil

}

func createPrimaryPartition(fdisk *FDISK, sizeBytes int) error {
	var mbr structures.MBR

	err := mbr.DeserializeMBR(fdisk.path)
	if err != nil {
		return err
	}
	if !mbr.CanFitAnotherDisk(sizeBytes) {
		return errors.New("no se puede crear una particion por falta de espacio")
	}

	// fmt.Println("\nMBR original: ")
	// mbr.PrintMBR()

	availablePartition, startPartition, indexPartition := mbr.GetFirstAvailablePartition()
	if availablePartition == nil {
		return errors.New("no hay partitciones disponibles")
	}

	// fmt.Println("\nParticion disponible:")
	// availablePartition.PrintPartition()

	availablePartition.CreatePartition(startPartition, sizeBytes, fdisk.typ, fdisk.fit, fdisk.name)

	// fmt.Println("\nParticion creada (modificada):")
	// availablePartition.PrintPartition()

	mbr.Mbr_partitions[indexPartition] = *availablePartition

	// fmt.Println("\nParticiones del MBR:")
	// mbr.PrintPartitions()

	err = mbr.SerializeMBR(fdisk.path)
	if err != nil {
		return err
	}

	return nil

}

func createExtendedPartittion(fdisk *FDISK, sizeBytes int) error {
	var mbr structures.MBR

	err := mbr.DeserializeMBR(fdisk.path)
	if err != nil {
		return err
	}

	if !mbr.CanFitAnotherDisk(sizeBytes) {
		return errors.New("no se puede crear una particion por falta de espacio")
	}

	// fmt.Println("\nMBR original: ")
	// mbr.PrintMBR()

	if mbr.IsThereExtendedPartition() {
		return errors.New("no se puede crear mas de 1 particion extendida por disco")
	}

	availablePartition, startPartition, indexPartition := mbr.GetFirstAvailablePartition()
	if availablePartition == nil {
		return errors.New("no hay partitciones disponibles")
	}

	availablePartition.CreatePartition(startPartition, sizeBytes, fdisk.typ, fdisk.fit, fdisk.name)

	// fmt.Println("\nParticion creada (modificada):")
	// availablePartition.PrintPartition()

	mbr.Mbr_partitions[indexPartition] = *availablePartition

	// fmt.Println("\n Particiones del MBR(actualizado): ")
	// mbr.PrintPartitions()

	err = mbr.SerializeMBR(fdisk.path)
	if err != nil {
		return err
	}

	return nil

}

func deletePartition(fdisk *FDISK) error {
	mbr := &structures.MBR{}
	err := mbr.DeserializeMBR(fdisk.path)
	var logicPartition bool
	if err != nil {
		return err
	}
	partition, indexPartition := mbr.GetPartitionByName(fdisk.name)
	if partition == nil {
		partition, err = mbr.GetExtendedPartition()
		if err != nil {
			return err
		}
		logicPartition = true
	}
	var partitionStart int32
	var partitionSize int32
	if !logicPartition {
		partitionStart = partition.Part_start
		partitionSize = partition.Part_size
		cleanPartition := &structures.PARTITION{
			Part_status: [1]byte{'N'}, Part_type: [1]byte{'N'}, Part_fit: [1]byte{'N'}, Part_start: -1, Part_size: -1, Part_name: [16]byte{'N'}, Part_correlative: -1, Part_id: [4]byte{'N'},
		}
		mbr.Mbr_partitions[indexPartition] = *cleanPartition
		err = mbr.SerializeMBR(fdisk.path)
		if err != nil {
			return err
		}
	}
	if fdisk.delete == "full" {
		err := FullDeletePartition(partitionStart, partitionSize, fdisk.path)
		if err != nil {
			return err
		}
	}
	return nil
}

func FullDeletePartition(offset, amountBytes int32, diskPath string) error {
	file, err := os.OpenFile(diskPath, os.O_WRONLY|os.O_CREATE, 0664)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Seek(int64(offset), io.SeekStart)
	if err != nil {
		return err
	}
	nullContent := make([]byte, amountBytes)
	_, err = file.Write(nullContent)
	if err != nil {
		return err
	}
	return nil
}

func addPartition(fdisk *FDISK) error {
	sizeBytes, err := utils.ConvertToBytes(fdisk.add, fdisk.unit)
	if err != nil {
		return err
	}
	if fdisk.add > 0 {
		return increasePartition(fdisk, sizeBytes)
	} else {
		return shrinkPartition(fdisk, -sizeBytes)
	}
}

func shrinkPartition(fdisk *FDISK, sizeBytes int) error {
	mbr := &structures.MBR{}
	err := mbr.DeserializeMBR(fdisk.path)
	var logicPartition bool
	if err != nil {
		return err
	}
	partition, indexPartition := mbr.GetPartitionByName(fdisk.name)
	if partition == nil {
		partition, err = mbr.GetExtendedPartition()
		if err != nil {
			return err
		}
		logicPartition = true
	}
	if !logicPartition {
		if sizeBytes > int(partition.Part_size) {
			return errors.New("no se puede quitar bytes a la particion dado que quedaria en negativo el size")
		}
		partition.Part_size = partition.Part_size - int32(sizeBytes)

		mbr.Mbr_partitions[indexPartition] = *partition
		err = mbr.SerializeMBR(fdisk.path)
		if err != nil {
			return err
		}
	}
	return nil
}

func increasePartition(fdisk *FDISK, sizeBytes int) error {
	mbr := &structures.MBR{}
	err := mbr.DeserializeMBR(fdisk.path)
	var logicPartition bool
	if err != nil {
		return err
	}
	partition, indexPartition := mbr.GetPartitionByName(fdisk.name)
	if partition == nil {
		partition, err = mbr.GetExtendedPartition()
		if err != nil {
			return err
		}
		logicPartition = true
	}
	if !logicPartition {
		outcome := isItPosibleToAdd(partition.Part_start+partition.Part_size, mbr, sizeBytes, indexPartition, mbr.Mbr_size)
		if !outcome {
			return errors.New("no hay suficiente espacio como para adicionar bytes a la particion")
		}
		partition.Part_size += int32(sizeBytes)
		mbr.Mbr_partitions[indexPartition] = *partition
		err = mbr.SerializeMBR(fdisk.path)
		if err != nil {
			return err
		}
	}
	return nil
}

func isItPosibleToAdd(partitionEnd int32, mbr *structures.MBR, amountOfBytes int, indexPartition int, diskEnd int32) bool {
	var positions []int
	for i, part := range mbr.Mbr_partitions {
		if indexPartition == i {
			continue
		}
		if partitionEnd < part.Part_start {
			positions = append(positions, int(part.Part_start))
		}
	}
	positions = append(positions, int(diskEnd))
	minPosition := getLowest(positions)
	return minPosition >= int(partitionEnd)+amountOfBytes

}

func getLowest(array []int) int {
	min := array[0]
	for _, value := range array {
		if value < min {
			min = value
		}
	}
	return min
}
