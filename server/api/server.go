package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"server/analyzer"
	"server/console"
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
