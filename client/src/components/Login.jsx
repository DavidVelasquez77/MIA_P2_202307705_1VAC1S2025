import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import './Login.css'

function Login() {
  const [formData, setFormData] = useState({
    idParticion: '',
    usuario: '',
    contraseña: ''
  })
  const [error, setError] = useState('')
  const [isSubmitting, setIsSubmitting] = useState(false)
  
  const navigate = useNavigate()
  const { login, isLoading } = useAuth()

  const handleChange = (e) => {
    const { name, value } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: value
    }))
    // Limpiar error cuando el usuario empiece a escribir
    if (error) setError('')
  }

  const handleSubmit = async (e) => {
    e.preventDefault()
    setIsSubmitting(true)
    setError('')

    try {
      const result = await login(formData.idParticion, formData.usuario, formData.contraseña)
      
      if (result.success) {
        // Login exitoso, redirigir a la consola
        navigate('/')
      } else {
        // Mostrar error de login
        setError(result.error || 'Error al iniciar sesión')
      }
    } catch (err) {
      setError('Error de conexión. Verifica que el servidor esté ejecutándose.')
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleBackToConsole = () => {
    navigate('/')
  }

  return (
    <div className="login-container">
      <div className="login-form">
        <div className="login-header">
          <h1>Iniciar Sesión</h1>
          <button className="back-button" onClick={handleBackToConsole}>
            Volver a Consola
          </button>
        </div>
        
        {error && (
          <div className="error-message">
            <span>⚠️ {error}</span>
          </div>
        )}
        
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="idParticion">ID Partición:</label>
            <input
              type="text"
              id="idParticion"
              name="idParticion"
              value={formData.idParticion}
              onChange={handleChange}
              placeholder="Ej: A105"
              required
              disabled={isSubmitting || isLoading}
            />
          </div>
          
          <div className="form-group">
            <label htmlFor="usuario">Usuario:</label>
            <input
              type="text"
              id="usuario"
              name="usuario"
              value={formData.usuario}
              onChange={handleChange}
              placeholder="Nombre de usuario"
              required
              disabled={isSubmitting || isLoading}
            />
          </div>
          
          <div className="form-group">
            <label htmlFor="contraseña">Contraseña:</label>
            <input
              type="password"
              id="contraseña"
              name="contraseña"
              value={formData.contraseña}
              onChange={handleChange}
              placeholder="Contraseña"
              required
              disabled={isSubmitting || isLoading}
            />
          </div>
          
          <button 
            type="submit" 
            className="login-button"
            disabled={isSubmitting || isLoading}
          >
            {isSubmitting || isLoading ? '🔄 Iniciando sesión...' : '🔐 Iniciar Sesión'}
          </button>
        </form>
        
        <div className="login-info">
          <p>💡 <strong>Nota:</strong> Debes tener usuarios creados en la partición especificada.</p>
          <p>📝 Usa los comandos <code>mkgrp</code> y <code>mkusr</code> en la consola para crear usuarios.</p>
        </div>
      </div>
    </div>
  )
}

export default Login
