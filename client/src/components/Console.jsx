import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import ApiService from '../services/api'
import './Console.css'

function Console() {
  const [commands, setCommands] = useState('')
  const [output, setOutput] = useState('')
  const [isExecuting, setIsExecuting] = useState(false)
  const [serverStatus, setServerStatus] = useState('checking')
  
  const navigate = useNavigate()

  useEffect(() => {
    checkServerStatus()
  }, [])

  const checkServerStatus = async () => {
    try {
      await ApiService.checkHealth()
      setServerStatus('connected')
      setOutput('✅ Conectado al servidor MIA\n')
    } catch (error) {
      setServerStatus('disconnected')
      setOutput('❌ Error: No se pudo conectar al servidor\n' +
               '🔧 Asegúrate de que el servidor esté ejecutándose con: go run main.go server\n')
    }
  }

  const handleExecute = async () => {
    if (!commands.trim() || serverStatus !== 'connected') return
    
    setIsExecuting(true)
    setOutput(prev => prev + '🔄 Ejecutando comandos...\n')
    
    try {
      const commandList = commands.split('\n').filter(cmd => cmd.trim())
      
      if (commandList.length === 1) {
        // Ejecutar comando individual
        const response = await ApiService.executeCommand(commandList[0])
        handleSingleResponse(response, commandList[0])
      } else {
        // Ejecutar comandos en lote
        const response = await ApiService.executeBatchCommands(commandList)
        handleBatchResponse(response)
      }
      
    } catch (error) {
      setOutput(prev => prev + `❌ Error de conexión: ${error.message}\n`)
    } finally {
      setIsExecuting(false)
    }
  }

  const handleSingleResponse = (response, command) => {
    let result = `\n📝 Comando: ${command}\n`
    
    if (response.success) {
      result += `✅ Éxito: ${response.message}\n`
      if (response.data) {
        result += `📄 Resultado: ${typeof response.data === 'string' ? response.data : JSON.stringify(response.data, null, 2)}\n`
      }
    } else {
      result += `❌ Error: ${response.error}\n`
    }
    
    result += '─'.repeat(80) + '\n'
    setOutput(prev => prev + result)
  }

  const handleBatchResponse = (response) => {
    let result = '\n📊 RESUMEN DE EJECUCIÓN EN LOTE\n'
    result += `📈 Total: ${response.summary.total} | ✅ Éxitos: ${response.summary.success} | ❌ Errores: ${response.summary.error}\n`
    result += '═'.repeat(80) + '\n'
    
    response.results.forEach((cmdResponse, index) => {
      const command = commands.split('\n')[index]?.trim()
      if (!command || command.startsWith('#')) return
      
      result += `\n[${index + 1}] ${command}\n`
      
      if (cmdResponse.success) {
        result += `    ✅ ${cmdResponse.message}\n`
        if (cmdResponse.data) {
          result += `    📄 ${typeof cmdResponse.data === 'string' ? cmdResponse.data : JSON.stringify(cmdResponse.data)}\n`
        }
      } else {
        result += `    ❌ ${cmdResponse.error}\n`
      }
    })
    
    result += '\n' + '═'.repeat(80) + '\n'
    setOutput(prev => prev + result)
  }

  const handleClear = () => {
    setCommands('')
    setOutput(serverStatus === 'connected' ? '✅ Conectado al servidor MIA\n' : '❌ Error: No se pudo conectar al servidor\n')
  }

  const handleGoToLogin = () => {
    navigate('/login')
  }

  const getServerStatusDisplay = () => {
    switch (serverStatus) {
      case 'connected':
        return '🟢 Conectado'
      case 'disconnected':
        return '🔴 Desconectado'
      default:
        return '🟡 Verificando...'
    }
  }

  return (
    <div className="console-container">
      <div className="console-header">
        <div className="header-left">
          <h1>Consola MIA</h1>
          <span className={`server-status ${serverStatus}`}>
            {getServerStatusDisplay()}
          </span>
        </div>
        <div className="header-right">
          <button className="refresh-button" onClick={checkServerStatus}>
            🔄 Reconectar
          </button>
          <button className="login-nav-button" onClick={handleGoToLogin}>
            Ir a Login
          </button>
        </div>
      </div>
      
      <div className="console-content">
        <div className="input-section">
          <label htmlFor="commands">Comandos a ejecutar:</label>
          <textarea
            id="commands"
            value={commands}
            onChange={(e) => setCommands(e.target.value)}
            placeholder="Ingresa los comandos, uno por línea...&#10;Ejemplo:&#10;mkdisk -size=3000 -unit=M -path=/home/disk1.dsk&#10;fdisk -size=300 -unit=M -path=/home/disk1.dsk -name=particion1&#10;mount -driveletter=A -name=particion1&#10;mkfs -id=A105&#10;mkusr -user=usuario1 -pwd=123456 -grp=usuarios"
            rows={10}
            disabled={isExecuting || serverStatus !== 'connected'}
          />
          
          <div className="button-group">
            <button 
              className="execute-button" 
              onClick={handleExecute}
              disabled={isExecuting || !commands.trim() || serverStatus !== 'connected'}
            >
              {isExecuting ? '⏳ Ejecutando...' : '▶️ Ejecutar Comandos'}
            </button>
            <button 
              className="clear-button" 
              onClick={handleClear}
              disabled={isExecuting}
            >
              🗑️ Limpiar
            </button>
          </div>
        </div>
        
        <div className="output-section">
          <label htmlFor="output">Salida del servidor:</label>
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
