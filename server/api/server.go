package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"server/analyzer"
	"server/console"
	"strings"
)

type CommandRequest struct {
	Command string `json:"command"`
}

type CommandResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
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

	// Procesar el comando
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
