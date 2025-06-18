import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import ApiService from '../services/api'
import './Console.css'

function Console() {
  const [commands, setCommands] = useState('')
  const [output, setOutput] = useState('')
  const [isExecuting, setIsExecuting] = useState(false)
  const [serverStatus, setServerStatus] = useState('checking')
  const [pendingCommand, setPendingCommand] = useState(null)
  const [userInput, setUserInput] = useState('')
  const [showInputDialog, setShowInputDialog] = useState(false)
  
  const { isAuthenticated, user, logout } = useAuth()
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
        await executeCommand(commandList[0])
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

  const executeCommand = async (command) => {
    const response = await ApiService.executeCommand(command)
    
    if (response.requiresInput) {
      setPendingCommand(response)
      setShowInputDialog(true)
      
      // Mostrar el comando y el mensaje de estado en la consola
      let result = `\n📝 Comando: ${command}\n`
      if (response.message) {
        result += `ℹ️ ${response.message}\n`
      }
      result += `❓ ${response.inputPrompt}\n`
      
      setOutput(prev => prev + result)
    } else {
      handleSingleResponse(response, command)
    }
  }

  const handleUserInputSubmit = async (inputValue) => {
    if (!pendingCommand) return
    
    setShowInputDialog(false)
    
    // Mostrar la respuesta del usuario en la consola
    const displayInput = pendingCommand.inputType === 'enter' ? '[ENTER]' : inputValue
    setOutput(prev => prev + `💬 Usuario respondió: ${displayInput}\n`)
    
    setUserInput('')
    
    try {
      const response = await ApiService.executeCommandWithInput(
        pendingCommand.pendingCommand, 
        inputValue
      )
      
      handleSingleResponse(response, pendingCommand.pendingCommand, false)
      
    } catch (error) {
      setOutput(prev => prev + `❌ Error: ${error.message}\n`)
    }
    
    setPendingCommand(null)
  }

  const handleInputDialogCancel = () => {
    setShowInputDialog(false)
    setUserInput('')
    setPendingCommand(null)
    setOutput(prev => prev + `❌ Comando cancelado por el usuario\n`)
  }

  const handleSingleResponse = (response, command, showCommand = true) => {
    let result = ''
    
    if (showCommand) {
      result += `\n📝 Comando: ${command}\n`
    }
    
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

  const handleLogout = async () => {
    try {
      await logout()
      setOutput(prev => prev + '👋 Sesión cerrada exitosamente\n')
    } catch (error) {
      setOutput(prev => prev + `❌ Error al cerrar sesión: ${error.message}\n`)
    }
  }

  const handleGoToFileSystem = () => {
    navigate('/filesystem')
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
          {isAuthenticated && user && (
            <div className="user-info">
              <span className="user-badge">
                👤 {user.usuario} @ {user.idParticion}
              </span>
            </div>
          )}
        </div>
        <div className="header-right">
          <button className="refresh-button" onClick={checkServerStatus}>
            🔄 Reconectar
          </button>
          {isAuthenticated ? (
            <>
              <button className="filesystem-button" onClick={handleGoToFileSystem}>
                📁 Explorador
              </button>
              <button className="logout-button" onClick={handleLogout}>
                🚪 Cerrar Sesión
              </button>
            </>
          ) : (
            <button className="login-nav-button" onClick={handleGoToLogin}>
              🔐 Iniciar Sesión
            </button>
          )}
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
      
      {showInputDialog && (
        <InputDialog
          prompt={pendingCommand?.inputPrompt}
          inputType={pendingCommand?.inputType}
          value={userInput}
          onChange={setUserInput}
          onSubmit={handleUserInputSubmit}
          onCancel={handleInputDialogCancel}
        />
      )}
    </div>
  )
}

function InputDialog({ prompt, inputType, value, onChange, onSubmit, onCancel }) {
  const handleSubmit = (e) => {
    e.preventDefault()
    if (inputType === 'yesno' && value.toLowerCase() !== 'y' && value.toLowerCase() !== 'n') {
      alert('Por favor ingresa "y" para sí o "n" para no')
      return
    }
    onSubmit(value)
  }

  const handleKeyPress = (e) => {
    if (inputType === 'enter' && e.key === 'Enter') {
      e.preventDefault()
      onSubmit('')
    }
  }

  return (
    <div className="input-dialog-overlay">
      <div className="input-dialog">
        <div className="input-dialog-header">
          <h3>🔍 Entrada requerida</h3>
        </div>
        
        <div className="input-dialog-body">
          <p><strong>{prompt}</strong></p>
          
          <form onSubmit={handleSubmit}>
            {inputType === 'enter' ? (
              <div className="enter-prompt">
                <span>⏸️ Presiona ENTER para continuar...</span>
                <input
                  type="text"
                  placeholder="Presiona Enter"
                  onKeyPress={handleKeyPress}
                  autoFocus
                  style={{ opacity: 0, position: 'absolute' }}
                />
              </div>
            ) : (
              <div className="text-input">
                <input
                  type="text"
                  value={value}
                  onChange={(e) => onChange(e.target.value)}
                  placeholder={inputType === 'yesno' ? 'Escribe "y" o "n"' : 'Ingresa tu respuesta'}
                  autoFocus
                  maxLength={inputType === 'yesno' ? 1 : undefined}
                />
                {inputType === 'yesno' && (
                  <div className="input-hint">
                    <small>💡 Escribe "y" para confirmar o "n" para cancelar</small>
                  </div>
                )}
              </div>
            )}
            
            <div className="input-dialog-buttons">
              <button type="submit" className="submit-button">
                {inputType === 'enter' ? '▶️ Continuar' : '📤 Enviar'}
              </button>
              <button type="button" onClick={onCancel} className="cancel-button">
                ❌ Cancelar
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  )
}

export default Console
