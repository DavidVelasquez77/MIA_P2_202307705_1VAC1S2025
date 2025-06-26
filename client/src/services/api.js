const API_BASE_URL = import.meta.env.VITE_API_URL
//http://localhost:8080/api
// import.meta.env.VITE_API_URL
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
      console.log(`🔄 Solicitando contenido - Partición: ${partitionId}, Ruta: ${path}`)
      
      const url = `${API_BASE_URL}/filesystem?partition=${encodeURIComponent(partitionId)}&path=${encodeURIComponent(path)}`
      console.log(`📡 URL completa: ${url}`)
      
      const response = await fetch(url, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
        },
        mode: 'cors',
      })

      console.log(`📊 Response status: ${response.status}`)
      console.log(`📊 Response headers:`, response.headers)

      if (!response.ok) {
        const errorText = await response.text()
        console.error('❌ Error response:', response.status, errorText)
        throw new Error(`HTTP ${response.status}: ${errorText}`)
      }

      const result = await response.json()
      console.log('✅ Respuesta del servidor:', result)
      return result
    } catch (error) {
      console.error('❌ Error en getFileSystem:', error)
      
      // Si es un error de red específico
      if (error.name === 'TypeError' && error.message.includes('fetch')) {
        throw new Error('Error de conexión: No se puede conectar al servidor. Verifica que esté ejecutándose.')
      }
      
      throw error
    }
  }

  async getFileContent(partitionId, filePath) {
    try {
      const response = await fetch(`${API_BASE_URL}/file-content?partition=${partitionId}&path=${encodeURIComponent(filePath)}`, {
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
      console.error('Error obteniendo contenido del archivo:', error)
      throw error
    }
  }
}

export default new ApiService()
