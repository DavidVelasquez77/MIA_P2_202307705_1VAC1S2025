package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"server/analyzer"
	"server/console"
	"server/stores"
	"server/structures"
	"server/utils"
	"strconv"
	"strings"
	"time"
)

type CommandRequest struct {
	Command string `json:"command"`
	Input   string `json:"input,omitempty"` // Para respuestas de usuario
}

type CommandResponse struct {
	Success        bool        `json:"success"`
	Message        string      `json:"message"`
	Data           interface{} `json:"data,omitempty"`
	Error          string      `json:"error,omitempty"`
	RequiresInput  bool        `json:"requiresInput,omitempty"`
	InputPrompt    string      `json:"inputPrompt,omitempty"`
	InputType      string      `json:"inputType,omitempty"` // "enter", "yesno"
	PendingCommand string      `json:"pendingCommand,omitempty"`
}

type BatchCommandRequest struct {
	Commands []string `json:"commands"`
}

type BatchCommandResponse struct {
	Success bool              `json:"success"`
	Results []CommandResponse `json:"results"`
	Summary map[string]int    `json:"summary"`
}

func StartServer(port string) {
	http.HandleFunc("/api/command", handleCommand)
	http.HandleFunc("/api/batch", handleBatchCommands)
	http.HandleFunc("/api/disks", handleGetDisks)
	http.HandleFunc("/api/partitions", handleGetPartitions)
	http.HandleFunc("/api/filesystem", handleGetFileSystem)
	http.HandleFunc("/api/file-content", handleGetFileContent)
	http.HandleFunc("/api/health", handleHealth)

	// Configurar CORS
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		// Responder con información básica para la raíz
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "running",
				"service": "MIA File System API",
				"message": "API funcionando correctamente",
				"time":    time.Now().Format("2006-01-02 15:04:05"),
			})
			return
		}
		http.NotFound(w, r)
	})

	console.PrintInfo(fmt.Sprintf("🚀 Servidor API iniciado en puerto %s", port))
	console.PrintInfo("📡 Endpoints disponibles:")
	console.PrintInfo("   GET / - Información básica del servidor")
	console.PrintInfo("   POST /api/command - Ejecutar comando individual")
	console.PrintInfo("   POST /api/batch - Ejecutar múltiples comandos")
	console.PrintInfo("   GET /api/disks - Obtener discos disponibles")
	console.PrintInfo("   GET /api/partitions?disk=<id> - Obtener particiones")
	console.PrintInfo("   GET /api/filesystem?partition=<id>&path=<path> - Obtener contenido")
	console.PrintInfo("   GET /api/file-content?partition=<id>&path=<path> - Obtener archivo")
	console.PrintInfo("   GET /api/health - Estado del servidor")
	console.PrintSeparator()

	serverAddr := "0.0.0.0:" + port
	console.PrintInfo(fmt.Sprintf("🔗 Servidor escuchando en %s", serverAddr))
	console.PrintInfo(fmt.Sprintf("🌍 Acceso externo: http://44.204.174.145:%s", port))
	console.PrintInfo(fmt.Sprintf("🏠 Acceso local: http://localhost:%s", port))
	console.PrintInfo("⚠️  Asegúrate de que el puerto esté abierto en el grupo de seguridad de AWS")

	console.PrintInfo("🔥 Iniciando servidor HTTP...")
	if err := http.ListenAndServe(serverAddr, nil); err != nil {
		console.PrintError(fmt.Sprintf("Error al iniciar servidor: %v", err))
		log.Fatal(err)
	}
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func handleCommand(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req CommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := CommandResponse{
			Success: false,
			Error:   "Error al decodificar el JSON: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Procesar comando directamente sin confirmaciones
	result, err := analyzer.Analyzer(req.Command)

	var response CommandResponse
	if err != nil {
		response = CommandResponse{
			Success: false,
			Error:   err.Error(),
		}
		console.PrintError(fmt.Sprintf("Error ejecutando comando '%s': %v", req.Command, err))
	} else {
		response = CommandResponse{
			Success: true,
			Message: "Comando ejecutado exitosamente",
			Data:    result,
		}
		console.PrintSuccess(fmt.Sprintf("Comando ejecutado: %s", req.Command))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleBatchCommands(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req BatchCommandRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := BatchCommandResponse{
			Success: false,
			Results: []CommandResponse{{
				Success: false,
				Error:   "Error al decodificar el JSON: " + err.Error(),
			}},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	var results []CommandResponse
	summary := map[string]int{"success": 0, "error": 0, "total": len(req.Commands)}

	console.PrintInfo(fmt.Sprintf("Ejecutando %d comandos en lote", len(req.Commands)))

	for i, command := range req.Commands {
		command = strings.TrimSpace(command)

		// Saltar comandos vacíos o comentarios
		if command == "" || strings.HasPrefix(command, "#") {
			continue
		}

		console.PrintCommand(fmt.Sprintf("[%d] %s", i+1, command))

		result, err := analyzer.Analyzer(command)

		var cmdResponse CommandResponse
		if err != nil {
			cmdResponse = CommandResponse{
				Success: false,
				Error:   err.Error(),
			}
			summary["error"]++
			console.PrintError(fmt.Sprintf("Error en comando %d: %v", i+1, err))
		} else {
			cmdResponse = CommandResponse{
				Success: true,
				Message: "Comando ejecutado exitosamente",
				Data:    result,
			}
			summary["success"]++
			console.PrintSuccess(fmt.Sprintf("Comando %d ejecutado correctamente", i+1))
		}

		results = append(results, cmdResponse)
	}

	response := BatchCommandResponse{
		Success: summary["error"] == 0,
		Results: results,
		Summary: summary,
	}

	console.PrintInfo(fmt.Sprintf("Lote completado: %d éxitos, %d errores", summary["success"], summary["error"]))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	response := map[string]interface{}{
		"status":  "healthy",
		"service": "MIA File System API",
		"version": "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetDisks(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Obtener discos reales del sistema
	disks := []map[string]interface{}{}

	// Debug: imprimir estado actual
	console.PrintInfo(fmt.Sprintf("🔍 Consultando discos cargados: %d discos encontrados", len(stores.LoadedDiskPaths)))

	for diskName, diskPath := range stores.LoadedDiskPaths {
		console.PrintInfo(fmt.Sprintf("  📀 Procesando disco: %s -> %s", diskName, diskPath))

		// Leer información del MBR
		mbr := &structures.MBR{}
		err := mbr.DeserializeMBR(diskPath)
		if err != nil {
			console.PrintError(fmt.Sprintf("  ❌ Error al leer MBR del disco %s: %v", diskName, err))
			continue // Saltar discos con errores
		}

		// Convertir tamaño a formato legible
		sizeInMB := float64(mbr.Mbr_size) / (1024 * 1024)
		sizeStr := fmt.Sprintf("%.1f MB", sizeInMB)

		disk := map[string]interface{}{
			"id":     diskName,
			"name":   diskName,
			"size":   sizeStr,
			"status": "Disponible",
			"path":   diskPath,
		}
		disks = append(disks, disk)
		console.PrintInfo(fmt.Sprintf("  ✅ Disco agregado a respuesta: %s", diskName))
	}

	console.PrintInfo(fmt.Sprintf("📊 Respuesta final: %d discos en la lista", len(disks)))

	response := map[string]interface{}{
		"success": true,
		"disks":   disks,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetPartitions(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	diskId := r.URL.Query().Get("disk")
	if diskId == "" {
		console.PrintError("Parámetro disk faltante en la solicitud")
		http.Error(w, "Parámetro disk requerido", http.StatusBadRequest)
		return
	}

	console.PrintInfo(fmt.Sprintf("🔍 Solicitud de particiones para disco: %s", diskId))

	// Debug: Mostrar estado actual de discos cargados
	console.PrintInfo(fmt.Sprintf("📊 Discos disponibles: %d", len(stores.LoadedDiskPaths)))
	for letter, path := range stores.LoadedDiskPaths {
		console.PrintInfo(fmt.Sprintf("  - %s: %s", letter, path))
	}

	// Obtener particiones reales del disco
	diskPath, exists := stores.LoadedDiskPaths[diskId]
	if !exists {
		console.PrintError(fmt.Sprintf("❌ Disco %s no encontrado en discos cargados", diskId))

		// Dar información detallada del error
		response := map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Disco %s no encontrado. Discos disponibles: %v", diskId, getAvailableDiskIds()),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	console.PrintInfo(fmt.Sprintf("✅ Disco encontrado: %s -> %s", diskId, diskPath))

	// Verificar que el archivo existe
	if _, err := os.Stat(diskPath); os.IsNotExist(err) {
		console.PrintError(fmt.Sprintf("❌ Archivo de disco no existe: %s", diskPath))
		response := map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("El archivo del disco %s no existe en %s", diskId, diskPath),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	mbr := &structures.MBR{}
	err := mbr.DeserializeMBR(diskPath)
	if err != nil {
		console.PrintError(fmt.Sprintf("❌ Error al leer MBR del disco %s: %v", diskId, err))
		response := map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Error al leer MBR del disco %s: %v", diskId, err),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	console.PrintInfo("📋 Leyendo particiones del MBR...")

	partitions := []map[string]interface{}{}
	partitionCount := 0

	for i, partition := range mbr.Mbr_partitions {
		console.PrintInfo(fmt.Sprintf("  Partición %d: tipo=%c, start=%d, size=%d",
			i, partition.Part_type[0], partition.Part_start, partition.Part_size))

		// Saltar particiones vacías o no utilizadas
		if partition.Part_type[0] == 'N' || partition.Part_start == -1 || partition.Part_size <= 0 {
			console.PrintInfo(fmt.Sprintf("    ⏭️ Saltando partición %d (no utilizada)", i))
			continue
		}

		partName := strings.TrimRight(string(partition.Part_name[:]), "\x00")
		partId := strings.TrimRight(string(partition.Part_id[:]), "\x00")

		// Verificar que tenga nombre válido
		if partName == "" {
			console.PrintInfo(fmt.Sprintf("    ⏭️ Saltando partición %d (sin nombre)", i))
			continue
		}

		var partType string
		switch partition.Part_type[0] {
		case 'P':
			partType = "Primaria"
		case 'E':
			partType = "Extendida"
		case 'L':
			partType = "Lógica"
		default:
			partType = fmt.Sprintf("Desconocida (%c)", partition.Part_type[0])
		}

		sizeInMB := float64(partition.Part_size) / (1024 * 1024)
		sizeStr := fmt.Sprintf("%.1f MB", sizeInMB)

		// Verificar si está montada
		mounted := partition.Part_status[0] == '1'

		part := map[string]interface{}{
			"id":      partId,
			"name":    partName,
			"type":    partType,
			"size":    sizeStr,
			"mounted": mounted,
			"start":   partition.Part_start,
			"rawSize": partition.Part_size,
		}
		partitions = append(partitions, part)
		partitionCount++

		console.PrintInfo(fmt.Sprintf("    ✅ Partición agregada: %s (%s, %s, montada: %v)",
			partName, partType, sizeStr, mounted))
	}

	console.PrintInfo(fmt.Sprintf("📊 Total de particiones procesadas: %d", partitionCount))

	response := map[string]interface{}{
		"success":    true,
		"partitions": partitions,
		"diskId":     diskId,
		"diskPath":   diskPath,
		"total":      partitionCount,
	}

	console.PrintInfo(fmt.Sprintf("✅ Respuesta enviada con %d particiones", len(partitions)))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Función auxiliar para obtener IDs de discos disponibles
func getAvailableDiskIds() []string {
	var ids []string
	for id := range stores.LoadedDiskPaths {
		ids = append(ids, id)
	}
	return ids
}

func handleGetFileSystem(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	partitionId := r.URL.Query().Get("partition")
	path := r.URL.Query().Get("path")

	if partitionId == "" || path == "" {
		http.Error(w, "Parámetros partition y path requeridos", http.StatusBadRequest)
		return
	}

	console.PrintInfo(fmt.Sprintf("📂 Solicitud filesystem - Partición: %s, Ruta: %s", partitionId, path))

	// Verificar si la partición existe en particiones montadas
	_, exists := stores.MountedPartitions[partitionId]
	if !exists {
		console.PrintError(fmt.Sprintf("Partición %s no está montada", partitionId))

		// Retornar error pero con estructura JSON válida
		response := map[string]interface{}{
			"success": false,
			"error":   "La partición no está montada. Use el comando mount para montarla.",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Obtener contenido real del sistema de archivos
	superBlock, _, diskPath, err := stores.GetMountedPartitionSuperblock(partitionId)
	if err != nil {
		console.PrintError(fmt.Sprintf("Error al obtener superblock: %v", err))

		response := map[string]interface{}{
			"success": false,
			"error":   "Error al obtener información de la partición: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Verificar que la partición tenga un sistema de archivos formateado
	if superBlock.S_magic != 0xEF53 {
		console.PrintWarning("Partición no formateada")

		response := map[string]interface{}{
			"success": false,
			"error":   "La partición no está formateada. Use el comando mkfs primero.",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	console.PrintInfo(fmt.Sprintf("✅ Superblock válido - Magic: 0x%X", superBlock.S_magic))

	// Para la raíz, siempre usar inodo 0
	var targetInodeIndex int32 = 0
	if path != "/" {
		// Navegar al directorio especificado
		inodeIndex, err := navigateToPath(superBlock, diskPath, path)
		if err != nil {
			console.PrintError(fmt.Sprintf("Error al navegar: %v", err))

			response := map[string]interface{}{
				"success": false,
				"error":   "Ruta no encontrada: " + err.Error(),
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
		targetInodeIndex = inodeIndex
	}

	console.PrintInfo(fmt.Sprintf("🎯 Leyendo inodo: %d", targetInodeIndex))

	// Obtener contenido del directorio usando el inodo encontrado
	folders, files, err := getDirectoryContentFromInode(superBlock, diskPath, targetInodeIndex, partitionId)
	if err != nil {
		console.PrintError(fmt.Sprintf("Error al leer contenido: %v", err))

		response := map[string]interface{}{
			"success": false,
			"error":   "Error al leer contenido del directorio: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	console.PrintInfo(fmt.Sprintf("📊 Resultado: %d carpetas, %d archivos", len(folders), len(files)))

	fileSystemContent := map[string]interface{}{
		"folders": folders,
		"files":   files,
	}

	response := map[string]interface{}{
		"success": true,
		"data":    fileSystemContent,
		"path":    path,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func navigateToPath(sb *structures.SuperBlock, diskPath string, path string) (int32, error) {
	// Normalizar la ruta
	if path == "/" {
		return 0, nil // Inodo raíz
	}

	// Limpiar la ruta y dividir en componentes
	path = strings.Trim(path, "/")
	pathComponents := strings.Split(path, "/")

	currentInodeIndex := int32(0) // Empezar desde la raíz

	// Navegar componente por componente
	for _, component := range pathComponents {
		if component == "" {
			continue
		}

		console.PrintInfo(fmt.Sprintf("Buscando componente: %s en inodo %d", component, currentInodeIndex))

		nextInodeIndex, err := findInodeInDirectory(sb, diskPath, currentInodeIndex, component)
		if err != nil {
			return -1, fmt.Errorf("no se encontró '%s' en la ruta: %v", component, err)
		}

		// Verificar que el inodo encontrado sea un directorio
		inode := &structures.Inode{}
		err = inode.Deserialize(diskPath, int64(sb.S_inode_start+(nextInodeIndex*sb.S_inode_size)))
		if err != nil {
			return -1, err
		}

		if inode.I_type[0] != '0' {
			return -1, fmt.Errorf("'%s' no es un directorio", component)
		}

		currentInodeIndex = nextInodeIndex
	}

	return currentInodeIndex, nil
}

func findInodeInDirectory(sb *structures.SuperBlock, diskPath string, dirInodeIndex int32, searchName string) (int32, error) {
	inode := &structures.Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+(dirInodeIndex*sb.S_inode_size)))
	if err != nil {
		return -1, fmt.Errorf("error al deserializar inodo %d: %v", dirInodeIndex, err)
	}

	// Verificar que sea un directorio
	if inode.I_type[0] != '0' {
		return -1, fmt.Errorf("el inodo %d no es un directorio (tipo: %c)", dirInodeIndex, inode.I_type[0])
	}

	console.PrintInfo(fmt.Sprintf("Buscando '%s' en directorio inodo %d", searchName, dirInodeIndex))

	// Buscar en todos los bloques del directorio
	for i, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break
		}

		console.PrintInfo(fmt.Sprintf("Revisando bloque %d (índice %d)", i, blockIndex))

		if i >= 14 {
			// Manejar bloques indirectos
			pointerBlock := &structures.PointerBlock{}
			err := pointerBlock.Deserialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {
				console.PrintError(fmt.Sprintf("Error al deserializar bloque indirecto: %v", err))
				continue
			}

			for _, ptrIndex := range pointerBlock.P_pointers {
				if ptrIndex == -1 {
					continue
				}

				folderBlock := &structures.FolderBlock{}
				err := folderBlock.Deserialize(diskPath, int64(sb.S_block_start+(ptrIndex*sb.S_block_size)))
				if err != nil {
					continue
				}

				inodeIndex, found := searchInFolderBlock(folderBlock, searchName)
				if found {
					console.PrintInfo(fmt.Sprintf("Encontrado '%s' en inodo %d", searchName, inodeIndex))
					return inodeIndex, nil
				}
			}
		} else {
			// Bloques directos
			folderBlock := &structures.FolderBlock{}
			err := folderBlock.Deserialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {
				console.PrintError(fmt.Sprintf("Error al deserializar bloque folder: %v", err))
				continue
			}

			// Debug: imprimir contenido del bloque
			console.PrintInfo(fmt.Sprintf("Contenido del bloque %d:", blockIndex))
			for j, content := range folderBlock.B_content {
				if content.B_inodo != -1 {
					name := strings.TrimRight(string(content.B_name[:]), "\x00")
					console.PrintInfo(fmt.Sprintf("  [%d] Nombre: '%s', Inodo: %d", j, name, content.B_inodo))
				}
			}

			inodeIndex, found := searchInFolderBlock(folderBlock, searchName)
			if found {
				console.PrintInfo(fmt.Sprintf("Encontrado '%s' en inodo %d", searchName, inodeIndex))
				return inodeIndex, nil
			}
		}
	}

	return -1, fmt.Errorf("no se encontró '%s' en el directorio inodo %d", searchName, dirInodeIndex)
}

func searchInFolderBlock(block *structures.FolderBlock, searchName string) (int32, bool) {
	for i := 0; i < len(block.B_content); i++ {
		content := block.B_content[i]
		if content.B_inodo == -1 {
			continue
		}

		contentName := strings.TrimRight(string(content.B_name[:]), "\x00")

		// Saltar entradas especiales pero NO saltar entradas con guión
		if contentName == "." || contentName == ".." || contentName == "" {
			continue
		}

		// Importante: no saltar entradas con "-" porque pueden ser archivos válidos
		if strings.EqualFold(contentName, searchName) {
			return content.B_inodo, true
		}
	}
	return -1, false
}

func getDirectoryContentFromInode(sb *structures.SuperBlock, diskPath string, inodeIndex int32, partitionId string) ([]map[string]interface{}, []map[string]interface{}, error) {
	console.PrintInfo(fmt.Sprintf("🔍 Leyendo inodo %d en posición: %d", inodeIndex, sb.S_inode_start+(inodeIndex*sb.S_inode_size)))

	inode := &structures.Inode{}
	err := inode.Deserialize(diskPath, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return nil, nil, fmt.Errorf("error al deserializar inodo %d: %v", inodeIndex, err)
	}

	console.PrintInfo(fmt.Sprintf("📋 Inodo %d - Tipo: %c, Bloques: %v", inodeIndex, inode.I_type[0], inode.I_block[:5]))

	// Verificar que sea un directorio
	if inode.I_type[0] != '0' {
		return nil, nil, fmt.Errorf("el inodo %d no es un directorio (tipo: %c)", inodeIndex, inode.I_type[0])
	}

	var folders []map[string]interface{}
	var files []map[string]interface{}

	// Recorrer todos los bloques del inodo
	for i, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break
		}

		console.PrintInfo(fmt.Sprintf("📦 Procesando bloque %d -> índice %d", i, blockIndex))

		if i >= 14 {
			// Manejar bloques indirectos
			console.PrintInfo("🔗 Procesando bloque indirecto")
			pointerBlock := &structures.PointerBlock{}
			err := pointerBlock.Deserialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {
				console.PrintError(fmt.Sprintf("Error en bloque indirecto: %v", err))
				continue
			}

			for j, ptrIndex := range pointerBlock.P_pointers {
				if ptrIndex == -1 {
					continue
				}

				console.PrintInfo(fmt.Sprintf("  📦 Sub-bloque %d -> índice %d", j, ptrIndex))

				folderBlock := &structures.FolderBlock{}
				err := folderBlock.Deserialize(diskPath, int64(sb.S_block_start+(ptrIndex*sb.S_block_size)))
				if err != nil {
					console.PrintError(fmt.Sprintf("Error al deserializar sub-bloque: %v", err))
					continue
				}

				f, fl := processDirectoryBlock(folderBlock, sb, diskPath, partitionId)
				folders = append(folders, f...)
				files = append(files, fl...)
			}
		} else {
			// Bloques directos
			folderBlock := &structures.FolderBlock{}
			err := folderBlock.Deserialize(diskPath, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {
				console.PrintError(fmt.Sprintf("Error al deserializar bloque folder %d: %v", blockIndex, err))
				continue
			}

			f, fl := processDirectoryBlock(folderBlock, sb, diskPath, partitionId)
			folders = append(folders, f...)
			files = append(files, fl...)
		}
	}

	console.PrintInfo(fmt.Sprintf("✅ Contenido procesado: %d carpetas, %d archivos", len(folders), len(files)))

	return folders, files, nil
}

func processDirectoryBlock(block *structures.FolderBlock, sb *structures.SuperBlock, diskPath string, partitionId string) ([]map[string]interface{}, []map[string]interface{}) {
	var folders []map[string]interface{}
	var files []map[string]interface{}

	console.PrintInfo("🗂️ Procesando bloque de directorio...")

	for i := 0; i < len(block.B_content); i++ {
		content := block.B_content[i]
		if content.B_inodo == -1 {
			continue
		}

		name := strings.TrimRight(string(content.B_name[:]), "\x00")

		// Saltar entradas especiales
		if name == "" || name == "." || name == ".." {
			continue
		}

		console.PrintInfo(fmt.Sprintf("📄 Entrada: '%s' -> inodo %d", name, content.B_inodo))

		// Obtener información del inodo
		itemInode := &structures.Inode{}
		err := itemInode.Deserialize(diskPath, int64(sb.S_inode_start+(content.B_inodo*sb.S_inode_size)))
		if err != nil {
			console.PrintError(fmt.Sprintf("Error al deserializar inodo %d para '%s': %v", content.B_inodo, name, err))
			continue
		}

		// Obtener permisos, propietario y grupo
		permissions := getPermissionString(string(itemInode.I_perm[:]))
		owner := getOwnerByIDSimple(itemInode.I_uid, partitionId)
		group := getGroupByIDSimple(itemInode.I_gid, partitionId)
		date := time.Unix(int64(itemInode.I_mtime), 0).Format("2006-01-02")

		if itemInode.I_type[0] == '0' {
			// Es una carpeta
			console.PrintInfo(fmt.Sprintf("📁 Agregando carpeta: %s", name))
			folders = append(folders, map[string]interface{}{
				"name":        name,
				"permissions": permissions,
				"owner":       owner,
				"group":       group,
				"size":        "4096",
				"date":        date,
			})
		} else if itemInode.I_type[0] == '1' {
			// Es un archivo
			console.PrintInfo(fmt.Sprintf("📄 Agregando archivo: %s (tamaño: %d)", name, itemInode.I_size))
			files = append(files, map[string]interface{}{
				"name":        name,
				"permissions": permissions,
				"owner":       owner,
				"group":       group,
				"size":        fmt.Sprintf("%d", itemInode.I_size),
				"date":        date,
			})
		}
	}

	return folders, files
}

func getPermissionString(perms string) string {
	if len(perms) < 3 {
		return "rwxrwxrwx"
	}

	var result string
	for _, perm := range perms {
		switch perm {
		case '0':
			result += "---"
		case '1':
			result += "--x"
		case '2':
			result += "-w-"
		case '3':
			result += "-wx"
		case '4':
			result += "r--"
		case '5':
			result += "r-x"
		case '6':
			result += "rw-"
		case '7':
			result += "rwx"
		default:
			result += "rwx"
		}
	}
	return result
}

func getOwnerByIDSimple(id int32, partitionId string) string {
	// Intentar obtener contenido del users.txt
	contentUsersTxt, err := getContetnUsersTxtSimple(partitionId)
	if err != nil {
		return "root"
	}

	strId := strconv.Itoa(int(id))
	contentMatrix := getContentMatrixUsers(contentUsersTxt)

	for _, row := range contentMatrix {
		if len(row) < 4 {
			continue
		}
		if row[0] == strId && row[1] == "U" {
			return row[3]
		}
	}
	return "root"
}

func getGroupByIDSimple(id int32, partitionId string) string {
	// Intentar obtener contenido del users.txt
	contentUsersTxt, err := getContetnUsersTxtSimple(partitionId)
	if err != nil {
		return "root"
	}

	strId := strconv.Itoa(int(id))
	contentMatrix := getContentMatrixUsers(contentUsersTxt)

	for _, row := range contentMatrix {
		if len(row) < 3 {
			continue
		}
		if row[0] == strId && row[1] == "G" {
			return row[2]
		}
	}
	return "root"
}

func getContetnUsersTxtSimple(partitionId string) (string, error) {
	partitionSuperblock, _, partitionPath, err := stores.GetMountedPartitionSuperblock(partitionId)
	if err != nil {
		return "", err
	}

	// Intentar obtener users.txt desde la raíz
	content, err := partitionSuperblock.ContentFromFile(partitionPath, 0, []string{}, "users.txt")
	if err != nil {
		return "", err
	}

	return content, nil
}

func processFileBlock(block *structures.FolderBlock, sb *structures.SuperBlock, diskPath string, partitionId string) ([]map[string]interface{}, []map[string]interface{}, error) {
	var folders []map[string]interface{}
	var files []map[string]interface{}

	for i := 2; i < len(block.B_content); i++ { // Saltar "." y ".."
		content := block.B_content[i]
		if content.B_inodo == -1 {
			continue
		}

		name := strings.TrimRight(string(content.B_name[:]), "\x00")
		if name == "" {
			continue
		}

		// Obtener información del inodo
		itemInode := &structures.Inode{}
		err := itemInode.Deserialize(diskPath, int64(sb.S_inode_start+(content.B_inodo*sb.S_inode_size)))
		if err != nil {
			continue
		}

		// Obtener permisos, propietario y grupo
		permissions := getPermissionString(string(itemInode.I_perm[:]))
		owner, _ := getOwnerByID(itemInode.I_uid, partitionId)
		group, _ := getGroupByID(itemInode.I_gid, partitionId)
		date := time.Unix(int64(itemInode.I_mtime), 0).Format("2006-01-02")

		if itemInode.I_type[0] == '0' {
			// Es una carpeta
			folders = append(folders, map[string]interface{}{
				"name":        name,
				"permissions": permissions,
				"owner":       owner,
				"group":       group,
				"size":        "4096",
				"date":        date,
			})
		} else {
			// Es un archivo
			files = append(files, map[string]interface{}{
				"name":        name,
				"permissions": permissions,
				"owner":       owner,
				"group":       group,
				"size":        fmt.Sprintf("%d", itemInode.I_size),
				"date":        date,
			})
		}
	}

	return folders, files, nil
}

func getOwnerByID(id int32, partitionId string) (string, error) {
	// Obtener contenido del users.txt
	contentUsersTxt, err := getContetnUsersTxt(partitionId)
	if err != nil {
		return "unknown", err
	}

	strId := fmt.Sprintf("%d", id)
	contentMatrix := getContentMatrixUsers(contentUsersTxt)

	for _, row := range contentMatrix {
		if len(row) < 4 {
			continue
		}
		if row[0] == strId && row[1] == "U" {
			return row[3], nil
		}
	}

	return "unknown", nil
}

func getGroupByID(id int32, partitionId string) (string, error) {
	// Obtener contenido del users.txt
	contentUsersTxt, err := getContetnUsersTxt(partitionId)
	if err != nil {
		return "unknown", err
	}

	strId := fmt.Sprintf("%d", id)
	contentMatrix := getContentMatrixUsers(contentUsersTxt)

	for _, row := range contentMatrix {
		if len(row) < 3 {
			continue
		}
		if row[0] == strId && row[1] == "G" {
			return row[2], nil
		}
	}

	return "unknown", nil
}

func getContetnUsersTxt(partitionId string) (string, error) {
	partitionSuperblock, _, partitionPath, err := stores.GetMountedPartitionSuperblock(partitionId)
	if err != nil {
		return "", err
	}

	parentDirs, destDir := utils.GetParentDirectories("/users.txt")
	content, err := partitionSuperblock.ContentFromFile(partitionPath, 0, parentDirs, destDir)
	if err != nil {
		return "", err
	}

	return content, nil
}

func getContentMatrixUsers(contentUsers string) [][]string {
	contentSplitedByEnters := strings.Split(contentUsers, "\n")
	if len(contentSplitedByEnters) > 0 && contentSplitedByEnters[len(contentSplitedByEnters)-1] == "" {
		contentSplitedByEnters = contentSplitedByEnters[:len(contentSplitedByEnters)-1]
	}

	var contentMatrix [][]string
	for _, value := range contentSplitedByEnters {
		if value != "" {
			contentMatrix = append(contentMatrix, strings.Split(value, ","))
		}
	}
	return contentMatrix
}

func handleGetFileContent(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	partitionId := r.URL.Query().Get("partition")
	filePath := r.URL.Query().Get("path")

	if partitionId == "" || filePath == "" {
		http.Error(w, "Parámetros partition y path requeridos", http.StatusBadRequest)
		return
	}

	console.PrintInfo(fmt.Sprintf("Obteniendo contenido de archivo: %s en partición: %s", filePath, partitionId))

	// Obtener contenido del archivo
	superBlock, _, diskPath, err := stores.GetMountedPartitionSuperblock(partitionId)
	if err != nil {
		console.PrintError(fmt.Sprintf("Error al obtener partición: %v", err))
		http.Error(w, "Error al obtener partición: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Verificar que la partición tenga un sistema de archivos formateado
	if superBlock.S_magic != 0xEF53 {
		console.PrintWarning("Partición no formateada")
		http.Error(w, "La partición no está formateada", http.StatusBadRequest)
		return
	}

	// Obtener contenido del archivo
	parentDirs, fileName := utils.GetParentDirectories(filePath)
	content, err := superBlock.ContentFromFile(diskPath, 0, parentDirs, fileName)
	if err != nil {
		console.PrintError(fmt.Sprintf("Error al leer archivo: %v", err))
		http.Error(w, "Error al leer archivo: "+err.Error(), http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"content": content,
		"path":    filePath,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
