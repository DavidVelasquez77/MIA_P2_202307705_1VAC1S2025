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
        
        if (response.disks.length === 0) {
          setError('No hay discos disponibles. Crea un disco primero usando el comando mkdisk.')
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
      
      // Obtener particiones del disco seleccionado usando el endpoint específico
      const response = await ApiService.getPartitions(disk.id)
      
      if (response.success) {
        setPartitions(response.partitions)
        
        if (response.partitions.length === 0) {
          setError('Este disco no tiene particiones. Crea una partición usando el comando fdisk.')
        }
      } else {
        setError('Error al cargar particiones del disco')
      }
    } catch (error) {
      setError('Error de conexión al cargar particiones')
      console.error('Error loading partitions:', error)
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

  const loadFileSystemContent = async (partitionId, path) => {
    try {
      // Usar el endpoint específico para obtener contenido del sistema de archivos
      const response = await ApiService.getFileSystem(partitionId, path)
      
      if (response.success && response.data) {
        setFileSystemData(response.data)
        setCurrentPath(path)
      } else {
        setError('Error al cargar contenido del directorio')
      }
    } catch (error) {
      setError('Error de conexión al cargar directorio')
      console.error('Error loading directory:', error)
    }
  }

  const handleBackToConsole = () => {
    navigate('/console')
  }

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const handleNavigateBack = () => {
    const pathParts = currentPath.split('/').filter(part => part !== '')
    pathParts.pop()
    const newPath = pathParts.length > 0 ? '/' + pathParts.join('/') : '/'
    loadFileSystemContent(selectedPartition.id, newPath)
  }

  const handleNavigateFolder = (folderName) => {
    const newPath = currentPath === '/' ? `/${folderName}` : `${currentPath}/${folderName}`
    loadFileSystemContent(selectedPartition.id, newPath)
  }

  if (!selectedDisk) {
    return (
      <div className="filesystem-viewer">
        <div className="viewer-header">
          <div className="header-left">
            <h1>Visualizador del Sistema de Archivos</h1>
            <p>Seleccione el disco que desea visualizar:</p>
          </div>
          <div className="header-right">
            <span className="user-info">👤 {user?.usuario}</span>
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
            <div className="disk-grid">
              {disks.map((disk) => (
                <div
                  key={disk.id}
                  className="disk-card"
                  onClick={() => handleDiskSelect(disk)}
                >
                  <div className="disk-icon">
                    <div className="disk-image">💾</div>
                  </div>
                  <div className="disk-info">
                    <h3>{disk.name}</h3>
                    <p>Tamaño: {disk.size}</p>
                    <p>Estado: {disk.status}</p>
                    <p>Letra: {disk.id}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        ) : !loading && !error && (
          <div className="no-data-message">
            <h3>📀 No hay discos disponibles</h3>
            <p>Para crear un disco, ve a la consola y usa el comando:</p>
            <code>mkdisk -size=1000 -unit=M</code>
          </div>
        )}

        {loading && (
          <div className="loading-overlay">
            <div className="loading-spinner">🔄 Cargando discos...</div>
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

        {fileSystemData && (
          <div className="file-list">
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
                {fileSystemData.folders.map((folder, index) => (
                  <tr 
                    key={`folder-${index}`} 
                    className="folder-row"
                    onClick={() => handleNavigateFolder(folder.name)}
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
                {fileSystemData.files.map((file, index) => (
                  <tr key={`file-${index}`} className="file-row">
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
