import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import './Login.css'

function Login() {
  const [formData, setFormData] = useState({
    idParticion: '',
    usuario: '',
    contraseña: ''
  })
  
  const navigate = useNavigate()

  const handleChange = (e) => {
    const { name, value } = e.target
    setFormData(prev => ({
      ...prev,
      [name]: value
    }))
  }

  const handleSubmit = (e) => {
    e.preventDefault()
    console.log('Datos de login:', formData)
    
    // Aquí puedes agregar la lógica para validar las credenciales
    navigate('/')
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
        
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="idParticion">ID Partición:</label>
            <input
              type="text"
              id="idParticion"
              name="idParticion"
              value={formData.idParticion}
              onChange={handleChange}
              required
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
              required
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
              required
            />
          </div>
          
          <button type="submit" className="login-button">
            Iniciar Sesión
          </button>
        </form>
      </div>
    </div>
  )
}

export default Login
