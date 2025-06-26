import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import ApiService from '../services/api'
import './FileSystemViewer.css'

function FileSystemViewer() {
  const [disks, setDisks] = useState([])
  const [selectedDisk, setSelectedDisk] = useState(null)
  const [partitions, setPartitions] = useState([])
  const [selectedPartition, setSelectedPartition] = useState(null)
  const [currentPath, setCurrentPath] = useState('/')
  const [fileSystemData, setFileSystemData] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [showFileContent, setShowFileContent] = useState(false)
  const [selectedFile, setSelectedFile] = useState(null)
  const [fileContent, setFileContent] = useState('')
  const [isRootUser, setIsRootUser] = useState(false) // Nueva variable para detectar root

  const { user, logout } = useAuth()
  const navigate = useNavigate()

  useEffect(() => {
    loadAvailableDisks()
  }, [])

  const loadAvailableDisks = async () => {
    try {
      setLoading(true)
      setError('')
      
      // Usar el endpoint específico para obtener discos
      const response = await ApiService.getDisks()
      
      if (response.success) {
        setDisks(response.disks)
        setIsRootUser(response.isRoot || false) // Obtener información si es root
        
        if (response.disks.length === 0) {
          if (response.isRoot) {
            setError('No hay discos disponibles en el sistema. Crea un disco primero usando el comando mkdisk.')
          } else {
            setError('No se encontró el disco asociado a tu partición. Contacta al administrador.')
          }
        }
      } else {
        setError('Error al cargar discos disponibles')
      }
    } catch (error) {
      setError('Error de conexión al cargar discos')
      console.error('Error loading disks:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleDiskSelect = async (disk) => {
    try {
      setLoading(true)
      setSelectedDisk(disk)
      setError('')
      
      console.log(`🔍 Solicitando particiones para disco: ${disk.id}`)
      
      // Obtener particiones del disco seleccionado usando el endpoint específico
      const response = await ApiService.getPartitions(disk.id)
      
      console.log('📦 Respuesta de particiones:', response)
      
      if (response && response.success) {
        console.log(`✅ Particiones recibidas: ${response.partitions?.length || 0}`)
        setPartitions(response.partitions || [])
        
        if (!response.partitions || response.partitions.length === 0) {
          setError(`Este disco (${disk.name}) no tiene particiones. Crea una partición usando el comando: fdisk -size=500 -path=${disk.path} -name=Particion1`)
        }
      } else {
        console.error('❌ Error en respuesta:', response)
        const errorMsg = response?.error || 'Error desconocido al cargar particiones'
        setError(`Error al cargar particiones: ${errorMsg}`)
      }
    } catch (error) {
      console.error('❌ Error de conexión:', error)
      setError(`Error de conexión al cargar particiones: ${error.message}`)
    } finally {
      setLoading(false)
    }
  }

  const handlePartitionSelect = async (partition) => {
    try {
      setLoading(true)
      setSelectedPartition(partition)
      setError('')
      
      if (!partition.mounted) {
        setError('La partición debe estar montada para explorar. Usa el comando mount para montarla.')
        return
      }
      
      // Cargar contenido del sistema de archivos
      await loadFileSystemContent(partition.id, '/')
    } catch (error) {
      setError('Error al acceder a la partición')
      console.error('Error accessing partition:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleNavigateFolder = async (folderName) => {
    try {
      setLoading(true)
      setError('')
      
      const newPath = currentPath === '/' ? `/${folderName}` : `${currentPath}/${folderName}`
      console.log(`Navegando a: ${newPath}`)
      
      await loadFileSystemContent(selectedPartition.id, newPath)
    } catch (error) {
      setError('Error al navegar a la carpeta: ' + error.message)
      console.error('Error navigating to folder:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleNavigateBack = async () => {
    if (currentPath === '/') return
    
    try {
      setLoading(true)
      setError('')
      
      const pathParts = currentPath.split('/').filter(part => part !== '')
      pathParts.pop()
      const newPath = pathParts.length === 0 ? '/' : '/' + pathParts.join('/')
      
      console.log(`Navegando hacia atrás a: ${newPath}`)
      
      await loadFileSystemContent(selectedPartition.id, newPath)
    } catch (error) {
      setError('Error al navegar hacia atrás: ' + error.message)
      console.error('Error navigating back:', error)
    } finally {
      setLoading(false)
    }
  }

  const loadFileSystemContent = async (partitionId, path) => {
    try {
      setError('')
      console.log(`🔄 Cargando contenido - Partición: ${partitionId}, Ruta: ${path}`)
      
      // Verificar que tenemos los parámetros necesarios
      if (!partitionId || !path) {
        throw new Error('Parámetros de partición o ruta faltantes')
      }
      
      // Usar el endpoint específico para obtener contenido del sistema de archivos
      const response = await ApiService.getFileSystem(partitionId, path)
      
      console.log('📦 Respuesta completa:', response)
      
      if (response && response.success && response.data) {
        console.log('✅ Datos válidos recibidos:', response.data)
        setFileSystemData(response.data)
        setCurrentPath(path)
        
        // Debug: Mostrar información en consola
        const totalFiles = (response.data.files?.length || 0)
        const totalFolders = (response.data.folders?.length || 0)
        console.log(`📊 Total: ${totalFolders} carpetas, ${totalFiles} archivos`)
        
      } else if (response && !response.success) {
        // Error del servidor pero respuesta válida
        setError(response.error || 'Error desconocido del servidor')
      } else {
        console.error('❌ Respuesta inválida del servidor:', response)
        setError('Respuesta inválida del servidor')
      }
    } catch (error) {
      console.error('❌ Error loading directory:', error)
      
      // Manejar diferentes tipos de errores
      let errorMessage = 'Error desconocido'
      
      if (error.message.includes('no se puede conectar')) {
        errorMessage = 'No se puede conectar al servidor. Verifica que esté ejecutándose en el puerto 8080.'
      } else if (error.message.includes('partición no está montada')) {
        errorMessage = `La partición ${partitionId} no está montada. Usa el comando "mount" en la consola.`
      } else if (error.message.includes('partición no está formateada')) {
        errorMessage = `La partición ${partitionId} no está formateada. Usa el comando "mkfs -id=${partitionId}" en la consola.`
      } else if (error.message.includes('404')) {
        errorMessage = 'Ruta no encontrada en el sistema de archivos.'
      } else if (error.message.includes('Failed to fetch')) {
        errorMessage = 'Error de conexión: No se puede conectar al servidor. Verifica que esté ejecutándose.'
      } else {
        errorMessage = `Error: ${error.message}`
      }
      
      setError(errorMessage)
    }
  }

  const handleBackToConsole = () => {
    navigate('/console')
  }

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const handleFileClick = async (fileName) => {
    try {
      setLoading(true)
      const filePath = currentPath === '/' ? `/${fileName}` : `${currentPath}/${fileName}`
      
      const response = await ApiService.getFileContent(selectedPartition.id, filePath)
      
      if (response.success) {
        setSelectedFile(fileName)
        setFileContent(response.content)
        setShowFileContent(true)
      } else {
        setError('Error al cargar contenido del archivo')
      }
    } catch (error) {
      setError('Error de conexión al cargar archivo')
      console.error('Error loading file:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleCloseFileContent = () => {
    setShowFileContent(false)
    setSelectedFile(null)
    setFileContent('')
  }

  if (!selectedDisk) {
    return (
      <div className="filesystem-viewer">
        <div className="viewer-header">
          <div className="header-left">
            <h1>Explorador del Sistema de Archivos</h1>
            {isRootUser ? (
              <div className="root-notice">
                <span className="root-badge">👑 ADMINISTRADOR</span>
                <p>📀 Tienes acceso a todos los discos del sistema</p>
              </div>
            ) : (
              <div className="user-notice">
                <p>📀 Mostrando el disco asociado a tu partición logueada</p>
                {user && (
                  <div className="user-disk-info">
                    <span className="info-badge">
                      👤 Usuario: {user.usuario} | 💾 Partición: {user.idParticion}
                    </span>
                  </div>
                )}
              </div>
            )}
          </div>
          <div className="header-right">
            <span className="user-info">
              {isRootUser ? '👑' : '👤'} {user?.usuario}
            </span>
            <button className="header-button" onClick={handleBackToConsole}>
              Consola
            </button>
            <button className="header-button logout" onClick={handleLogout}>
              Cerrar Sesión
            </button>
          </div>
        </div>

        {error && (
          <div className="error-message">
            <span>⚠️ {error}</span>
          </div>
        )}

        {disks.length > 0 ? (
          <div className="disk-selection">
            {isRootUser ? (
              <div className="admin-disk-notice">
                <h3>👑 Acceso de Administrador</h3>
                <p>Como administrador, puedes acceder a todos los discos del sistema</p>
                <p>📊 Total de discos disponibles: {disks.length}</p>
              </div>
            ) : (
              <div className="user-disk-notice">
                <h3>🔒 Acceso de Usuario</h3>
                <p>Solo puedes acceder al disco que contiene tu partición logueada ({user?.idParticion})</p>
              </div>
            )}
            <div className="disk-grid">
              {disks.map((disk) => (
                <div
                  key={disk.id}
                  className={`disk-card ${isRootUser ? 'admin-disk' : 'user-disk'}`}
                  onClick={() => handleDiskSelect(disk)}
                >
                  <div className="disk-icon">
                    <div className="disk-image">💾</div>
                    {isRootUser ? (
                      <div className="admin-disk-badge">👑 Admin</div>
                    ) : (
                      <div className="user-disk-badge">🔒 Tu Disco</div>
                    )}
                  </div>
                  <div className="disk-info">
                    <h3>{disk.name}</h3>
                    <p>Tamaño: {disk.size}</p>
                    <p>Estado: {disk.status}</p>
                    <p>Letra: {disk.id}</p>
                    {!isRootUser && (
                      <p className="user-partition">📁 Partición: {user?.idParticion}</p>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        ) : !loading && !error && (
          <div className="no-data-message">
            {isRootUser ? (
              <>
                <h3>📀 No hay discos en el sistema</h3>
                <p>Como administrador, no hay discos disponibles actualmente</p>
                <p>Para crear un disco, ve a la consola y usa el comando:</p>
                <code>mkdisk -size=1000 -unit=M</code>
              </>
            ) : (
              <>
                <h3>❌ No se encontró tu disco</h3>
                <p>No se pudo encontrar el disco asociado a tu partición {user?.idParticion}</p>
                <p>💡 Esto puede suceder si:</p>
                <ul>
                  <li>El disco fue desmontado después del login</li>
                  <li>Hay un problema con la configuración del sistema</li>
                  <li>La partición no está correctamente asociada</li>
                </ul>
                <p>🔧 Intenta hacer logout y login nuevamente</p>
              </>
            )}
          </div>
        )}

        {loading && (
          <div className="loading-overlay">
            <div className="loading-spinner">
              🔄 {isRootUser ? 'Cargando discos del sistema...' : 'Cargando tu disco...'}
            </div>
          </div>
        )}
      </div>
    )
  }

  if (!selectedPartition) {
    return (
      <div className="filesystem-viewer">
        <div className="viewer-header">
          <div className="header-left">
            <h1>Seleccionar Partición</h1>
            <p>Disco seleccionado: {selectedDisk.name}</p>
          </div>
          <div className="header-right">
            <button className="header-button" onClick={() => setSelectedDisk(null)}>
              ← Volver a Discos
            </button>
            <button className="header-button" onClick={handleLogout}>
              Cerrar Sesión
            </button>
          </div>
        </div>

        {error && (
          <div className="error-message">
            <span>⚠️ {error}</span>
          </div>
        )}

        <div className="partition-selection">
          <div className="partition-list">
            {partitions.map((partition) => (
              <div
                key={partition.id}
                className={`partition-card ${!partition.mounted ? 'disabled' : ''}`}
                onClick={() => partition.mounted && handlePartitionSelect(partition)}
              >
                <div className="partition-info">
                  <h3>{partition.name}</h3>
                  <p>Tipo: {partition.type}</p>
                  <p>Tamaño: {partition.size}</p>
                  <p>Estado: {partition.mounted ? 'Montada' : 'No montada'}</p>
                </div>
                <div className="partition-status">
                  {partition.mounted ? '✅' : '❌'}
                </div>
              </div>
            ))}
          </div>
        </div>

        {loading && (
          <div className="loading-overlay">
            <div className="loading-spinner">🔄 Cargando...</div>
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="filesystem-viewer">
      <div className="viewer-header">
        <div className="header-left">
          <h1>Explorador de Archivos</h1>
          <div className="breadcrumb">
            <span>Disco: {selectedDisk.name}</span>
            <span>→</span>
            <span>Partición: {selectedPartition.name}</span>
            <span>→</span>
            <span>Ruta: {currentPath}</span>
          </div>
        </div>
        <div className="header-right">
          <button className="header-button" onClick={() => setSelectedPartition(null)}>
            ← Particiones
          </button>
          <button className="header-button" onClick={handleBackToConsole}>
            Consola
          </button>
          <button className="header-button logout" onClick={handleLogout}>
            Cerrar Sesión
          </button>
        </div>
      </div>

      <div className="filesystem-content">
        <div className="navigation-bar">
          <button 
            className="nav-button"
            onClick={handleNavigateBack}
            disabled={currentPath === '/'}
          >
            ← Atrás
          </button>
          <span className="current-path">{currentPath}</span>
        </div>

        {error && (
          <div className="error-message">
            <span>⚠️ {error}</span>
          </div>
        )}

        {fileSystemData && !showFileContent && (
          <div className="file-list">
            <div className="file-list-header">
              <h4>📁 Contenido del directorio: {currentPath}</h4>
              <span>
                {(fileSystemData.folders?.length || 0) + (fileSystemData.files?.length || 0)} elementos
              </span>
            </div>
            
            {/* Mostrar mensaje si no hay contenido */}
            {(!fileSystemData.files || fileSystemData.files.length === 0) && 
             (!fileSystemData.folders || fileSystemData.folders.length === 0) ? (
              <div className="empty-directory-message">
                <h3>📭 Directorio vacío</h3>
                <p>Este directorio no contiene archivos ni carpetas.</p>
                <p>💡 <strong>Sugerencias:</strong></p>
                <ul>
                  <li>Crea carpetas con: <code>mkdir -path=/ruta/carpeta</code></li>
                  <li>Crea archivos con: <code>mkfile -path=/ruta/archivo.txt -size=100</code></li>
                  <li>Si acabas de formatear, el archivo users.txt debería estar aquí</li>
                </ul>
              </div>
            ) : (
              <table className="file-table">
                <thead>
                  <tr>
                    <th>Nombre</th>
                    <th>Tipo</th>
                    <th>Permisos</th>
                    <th>Propietario</th>
                    <th>Grupo</th>
                    <th>Tamaño</th>
                    <th>Fecha</th>
                  </tr>
                </thead>
                <tbody>
                  {fileSystemData.folders && fileSystemData.folders.map((folder, index) => (
                    <tr 
                      key={`folder-${index}`} 
                      className="folder-row"
                      onClick={() => handleNavigateFolder(folder.name)}
                      title={`Doble clic para entrar a ${folder.name}`}
                    >
                      <td>
                        <span className="file-icon">📁</span>
                        {folder.name}
                      </td>
                      <td>Carpeta</td>
                      <td>{folder.permissions}</td>
                      <td>{folder.owner}</td>
                      <td>{folder.group}</td>
                      <td>{folder.size}</td>
                      <td>{folder.date}</td>
                    </tr>
                  ))}
                  {fileSystemData.files && fileSystemData.files.map((file, index) => (
                    <tr 
                      key={`file-${index}`} 
                      className="file-row"
                      onClick={() => handleFileClick(file.name)}
                      title={`Clic para ver contenido de ${file.name}`}
                    >
                      <td>
                        <span className="file-icon">📄</span>
                        {file.name}
                      </td>
                      <td>Archivo</td>
                      <td>{file.permissions}</td>
                      <td>{file.owner}</td>
                      <td>{file.group}</td>
                      <td>{file.size} bytes</td>
                      <td>{file.date}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        )}

        {showFileContent && (
          <div className="file-content-viewer">
            <div className="file-content-header">
              <h3>📄 Contenido de: {selectedFile}</h3>
              <button className="close-file-button" onClick={handleCloseFileContent}>
                ❌ Cerrar
              </button>
            </div>
            <div className="file-content-body">
              <pre className="file-content-text">{fileContent}</pre>
            </div>
          </div>
        )}

        {loading && (
          <div className="loading-overlay">
            <div className="loading-spinner">🔄 Cargando...</div>
          </div>
        )}
      </div>
    </div>
  )
}

export default FileSystemViewer
