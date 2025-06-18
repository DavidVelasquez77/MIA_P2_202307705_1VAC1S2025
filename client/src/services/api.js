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

  async login(idParticion, usuario, contraseña) {
    try {
      const loginCommand = `login -user=${usuario} -pass=${contraseña} -id=${idParticion}`
      
      const response = await fetch(`${API_BASE_URL}/command`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ command: loginCommand }),
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const result = await response.json()
      
      if (result.success) {
        // Login exitoso
        return {
          success: true,
          message: result.data,
          userData: {
            idParticion,
            usuario,
            isLoggedIn: true
          }
        }
      } else {
        // Login fallido
        return {
          success: false,
          error: result.error
        }
      }
    } catch (error) {
      console.error('Error en login:', error)
      return {
        success: false,
        error: 'Error de conexión con el servidor'
      }
    }
  }

  async logout() {
    try {
      const response = await fetch(`${API_BASE_URL}/command`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ command: 'logout' }),
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error('Error en logout:', error)
      throw error
    }
  }

  async getDisks() {
    try {
      const response = await fetch(`${API_BASE_URL}/disks`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error('Error obteniendo discos:', error)
      throw error
    }
  }

  async getPartitions(diskId) {
    try {
      const response = await fetch(`${API_BASE_URL}/partitions?disk=${diskId}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error('Error obteniendo particiones:', error)
      throw error
    }
  }

  async getFileSystem(partitionId, path) {
    try {
      const response = await fetch(`${API_BASE_URL}/filesystem?partition=${partitionId}&path=${encodeURIComponent(path)}`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
      })

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      return await response.json()
    } catch (error) {
      console.error('Error obteniendo sistema de archivos:', error)
      throw error
    }
  }
}

export default new ApiService()
