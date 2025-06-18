package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"server/analyzer"
	"server/console"
	"server/stores"
	"server/structures"
	"strings"
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
	http.HandleFunc("/api/health", handleHealth)

	// Configurar CORS
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w)
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	})

	console.PrintInfo(fmt.Sprintf("🚀 Servidor API iniciado en puerto %s", port))
	console.PrintInfo("📡 Endpoints disponibles:")
	console.PrintInfo("   POST /api/command - Ejecutar comando individual")
	console.PrintInfo("   POST /api/batch - Ejecutar múltiples comandos")
	console.PrintInfo("   GET /api/health - Estado del servidor")
	console.PrintSeparator()

	log.Fatal(http.ListenAndServe(":"+port, nil))
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

	// Verificar si es un comando que requiere confirmación
	if requiresConfirmation(req.Command) && req.Input == "" {
		response := handleInteractiveCommand(req.Command)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Procesar comando con input si es necesario
	result, err := processCommandWithInput(req.Command, req.Input)

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

func requiresConfirmation(command string) bool {
	lowerCmd := strings.ToLower(strings.TrimSpace(command))
	return strings.HasPrefix(lowerCmd, "pause") ||
		(strings.Contains(lowerCmd, "fdisk") && strings.Contains(lowerCmd, "-delete"))
}

func handleInteractiveCommand(command string) CommandResponse {
	lowerCmd := strings.ToLower(strings.TrimSpace(command))

	if strings.HasPrefix(lowerCmd, "pause") {
		return CommandResponse{
			RequiresInput:  true,
			InputPrompt:    "Presiona ENTER para continuar...",
			InputType:      "enter",
			PendingCommand: command,
			Message:        "⏸️ PAUSE: Esperando confirmación del usuario",
		}
	}

	if strings.Contains(lowerCmd, "fdisk") && strings.Contains(lowerCmd, "-delete") {
		return CommandResponse{
			RequiresInput:  true,
			InputPrompt:    "Desea confirmar la ejecucion del delete? [y/n]:",
			InputType:      "yesno",
			PendingCommand: command,
			Message:        "⚠️ FDISK DELETE: Esperando confirmación del usuario",
		}
	}

	return CommandResponse{
		Success: false,
		Error:   "Comando no reconocido como interactivo",
	}
}

func processCommandWithInput(command, input string) (interface{}, error) {
	lowerCmd := strings.ToLower(strings.TrimSpace(command))

	if strings.HasPrefix(lowerCmd, "pause") {
		// Para pause, cualquier input es válido (incluso vacío)
		return analyzer.AnalyzerWithInput(command, "")
	}

	if strings.Contains(lowerCmd, "fdisk") && strings.Contains(lowerCmd, "-delete") {
		// Para fdisk delete, necesitamos y/n
		lowerInput := strings.ToLower(strings.TrimSpace(input))
		if lowerInput != "y" && lowerInput != "n" {
			return nil, errors.New("respuesta inválida. Use 'y' para sí o 'n' para no")
		}
		return analyzer.AnalyzerWithInput(command, lowerInput)
	}

	// Para comandos normales
	return analyzer.Analyzer(command)
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

	for diskName, diskPath := range stores.LoadedDiskPaths {
		// Leer información del MBR
		mbr := &structures.MBR{}
		err := mbr.DeserializeMBR(diskPath)
		if err != nil {
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
	}

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
		http.Error(w, "Parámetro disk requerido", http.StatusBadRequest)
		return
	}

	// Obtener particiones reales del disco
	diskPath, exists := stores.LoadedDiskPaths[diskId]
	if !exists {
		http.Error(w, "Disco no encontrado", http.StatusNotFound)
		return
	}

	mbr := &structures.MBR{}
	err := mbr.DeserializeMBR(diskPath)
	if err != nil {
		http.Error(w, "Error al leer MBR", http.StatusInternalServerError)
		return
	}

	partitions := []map[string]interface{}{}

	for _, partition := range mbr.Mbr_partitions {
		// Saltar particiones vacías
		if partition.Part_type[0] == 'N' || partition.Part_start == -1 {
			continue
		}

		partName := strings.TrimRight(string(partition.Part_name[:]), "\x00")
		partId := strings.TrimRight(string(partition.Part_id[:]), "\x00")

		var partType string
		switch partition.Part_type[0] {
		case 'P':
			partType = "Primaria"
		case 'E':
			partType = "Extendida"
		case 'L':
			partType = "Lógica"
		default:
			partType = "Desconocida"
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
		}
		partitions = append(partitions, part)
	}

	response := map[string]interface{}{
		"success":    true,
		"partitions": partitions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func handleGetFileSystem(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

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

	// Simular contenido del sistema de archivos
	fileSystemContent := map[string]interface{}{
		"folders": []map[string]interface{}{
			{"name": "users", "permissions": "rwxr-xr-x", "owner": "root", "group": "root", "size": "4096", "date": "2024-01-15"},
			{"name": "documents", "permissions": "rwxr-xr-x", "owner": "admin", "group": "users", "size": "4096", "date": "2024-01-10"},
			{"name": "temp", "permissions": "rwxrwxrwx", "owner": "root", "group": "root", "size": "4096", "date": "2024-01-12"},
		},
		"files": []map[string]interface{}{
			{"name": "users.txt", "permissions": "rw-r--r--", "owner": "root", "group": "root", "size": "245", "date": "2024-01-15"},
			{"name": "config.conf", "permissions": "rw-r--r--", "owner": "admin", "group": "users", "size": "1024", "date": "2024-01-14"},
			{"name": "readme.txt", "permissions": "rw-r--r--", "owner": "root", "group": "root", "size": "512", "date": "2024-01-13"},
		},
	}

	response := map[string]interface{}{
		"success": true,
		"data":    fileSystemContent,
		"path":    path,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
