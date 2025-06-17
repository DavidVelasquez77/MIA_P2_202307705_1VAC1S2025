import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import './Console.css'

function Console() {
  const [commands, setCommands] = useState('')
  const [output, setOutput] = useState('')
  const [isExecuting, setIsExecuting] = useState(false)
  
  const navigate = useNavigate()

  const handleExecute = async () => {
    if (!commands.trim()) return
    
    setIsExecuting(true)
    setOutput('Ejecutando comandos...\n')
    
    try {
      // Simular ejecución de comandos
      // Aquí puedes agregar la lógica para enviar los comandos al servidor
      const commandList = commands.split('\n').filter(cmd => cmd.trim())
      
      let result = 'Resultados de ejecución:\n'
      commandList.forEach((cmd, index) => {
        result += `[${index + 1}] ${cmd}\n`
        result += `    ✓ Comando ejecutado exitosamente\n`
      })
      
      setTimeout(() => {
        setOutput(result)
        setIsExecuting(false)
      }, 1000)
      
    } catch (error) {
      setOutput(`Error: ${error.message}`)
      setIsExecuting(false)
    }
  }

  const handleClear = () => {
    setCommands('')
    setOutput('')
  }

  const handleGoToLogin = () => {
    navigate('/login')
  }

  return (
    <div className="console-container">
      <div className="console-header">
        <h1>Consola de Comandos</h1>
        <button className="login-nav-button" onClick={handleGoToLogin}>
          Ir a Login
        </button>
      </div>
      
      <div className="console-content">
        <div className="input-section">
          <label htmlFor="commands">Comandos a ejecutar:</label>
          <textarea
            id="commands"
            value={commands}
            onChange={(e) => setCommands(e.target.value)}
            placeholder="Ingresa los comandos, uno por línea...&#10;Ejemplo:&#10;mkdisk -size=3000 -unit=M -path=/home/disk1.dsk&#10;fdisk -size=300 -unit=M -path=/home/disk1.dsk -name=particion1&#10;mkusr -user=usuario1 -pwd=123456 -grp=1"
            rows={10}
            disabled={isExecuting}
          />
          
          <div className="button-group">
            <button 
              className="execute-button" 
              onClick={handleExecute}
              disabled={isExecuting || !commands.trim()}
            >
              {isExecuting ? 'Ejecutando...' : 'Ejecutar Comandos'}
            </button>
            <button 
              className="clear-button" 
              onClick={handleClear}
              disabled={isExecuting}
            >
              Limpiar
            </button>
          </div>
        </div>
        
        <div className="output-section">
          <label htmlFor="output">Salida:</label>
          <textarea
            id="output"
            value={output}
            readOnly
            placeholder="La salida de los comandos aparecerá aquí..."
            rows={15}
          />
        </div>
      </div>
    </div>
  )
}

export default Console
