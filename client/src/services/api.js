const API_BASE_URL = 'http://localhost:8080/api'

class ApiService {
  async executeCommand(command) {
    try {
      const response = await fetch(`${API_BASE_URL}/command`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ command }),
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error('Error ejecutando comando:', error)
      throw error
    }
  }

  async executeCommandWithInput(command, input) {
    try {
      const response = await fetch(`${API_BASE_URL}/command`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ command, input }),
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error('Error ejecutando comando con input:', error)
      throw error
    }
  }

  async executeBatchCommands(commands) {
    try {
      const response = await fetch(`${API_BASE_URL}/batch`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ commands }),
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error('Error ejecutando comandos en lote:', error)
      throw error
    }
  }

  async checkHealth() {
    try {
      const response = await fetch(`${API_BASE_URL}/health`)
      return await response.json()
    } catch (error) {
      console.error('Error verificando estado del servidor:', error)
      throw error
    }
  }
}

export default new ApiService()
