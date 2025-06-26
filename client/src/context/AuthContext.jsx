import { createContext, useContext, useState, useEffect } from 'react'
import ApiService from '../services/api'

const AuthContext = createContext()

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth debe ser usado dentro de AuthProvider')
  }
  return context
}

export function AuthProvider({ children }) {
  const [isAuthenticated, setIsAuthenticated] = useState(false)
  const [user, setUser] = useState(null)
  const [isLoading, setIsLoading] = useState(false)

  // Verificar si hay una sesión guardada al cargar la aplicación
  useEffect(() => {
    const savedUser = localStorage.getItem('mia_user')
    if (savedUser) {
      try {
        const userData = JSON.parse(savedUser)
        setUser(userData)
        setIsAuthenticated(true)
      } catch (error) {
        // Si hay error al parsear, limpiar localStorage
        localStorage.removeItem('mia_user')
      }
    }
  }, [])

  const login = async (idParticion, usuario, contraseña) => {
    setIsLoading(true)
    
    try {
      // Crear comando de login
      const loginCommand = `login -id=${idParticion} -user=${usuario} -pass=${contraseña}`
      const result = await ApiService.executeCommand(loginCommand)
      
      if (result.success) {
        // Crear datos del usuario con información adicional
        const userData = {
          idParticion,
          usuario,
          loginTime: new Date().toISOString()
        }
        
        setUser(userData)
        setIsAuthenticated(true)
        
        // Guardar sesión en localStorage
        localStorage.setItem('mia_user', JSON.stringify(userData))
        
        return { 
          success: true, 
          message: result.data || `Usuario ${usuario} logueado exitosamente`,
          userData 
        }
      } else {
        return { success: false, error: result.error }
      }
    } catch (error) {
      return { 
        success: false, 
        error: 'Error de conexión. Verifica que el servidor esté ejecutándose.' 
      }
    } finally {
      setIsLoading(false)
    }
  }

  const logout = async () => {
    setIsLoading(true)
    
    try {
      // Intentar logout en el servidor solo si hay usuario autenticado
      if (isAuthenticated) {
        await ApiService.executeCommand('logout')
      }
    } catch (error) {
      console.error('Error al hacer logout en el servidor:', error)
    } finally {
      // Limpiar estado local independientemente del resultado del servidor
      setUser(null)
      setIsAuthenticated(false)
      localStorage.removeItem('mia_user')
      setIsLoading(false)
    }
  }

  const value = {
    isAuthenticated,
    user,
    isLoading,
    login,
    logout
  }

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  )
}
