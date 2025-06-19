import { useState, useEffect } from 'react'
import { useMusic } from '../context/MusicContext'
import './MusicPlayer.css'

function MusicPlayer() {
  const { 
    isPlaying, 
    volume, 
    currentTime, 
    duration, 
    isLoaded,
    togglePlay, 
    changeVolume, 
    seek 
  } = useMusic()
  
  const [showControls, setShowControls] = useState(false)
  const [loadingState, setLoadingState] = useState('loading')

  useEffect(() => {
    if (isLoaded) {
      setLoadingState('loaded')
    } else {
      setLoadingState('loading')
    }
  }, [isLoaded])

  const formatTime = (time) => {
    if (!time || isNaN(time)) return '0:00'
    const minutes = Math.floor(time / 60)
    const seconds = Math.floor(time % 60)
    return `${minutes}:${seconds.toString().padStart(2, '0')}`
  }

  const handleSeek = (e) => {
    if (!isLoaded || !duration) return
    
    const rect = e.currentTarget.getBoundingClientRect()
    const percent = (e.clientX - rect.left) / rect.width
    const newTime = percent * duration
    seek(newTime)
  }

  const progressPercent = duration ? (currentTime / duration) * 100 : 0

  const handlePlayToggle = () => {
    if (!isLoaded) {
      console.warn('Audio aún no está cargado')
      return
    }
    togglePlay()
  }

  return (
    <div className={`music-player ${showControls ? 'expanded' : ''}`}>
      {/* Botón principal de música */}
      <button 
        className={`music-toggle-button ${loadingState}`}
        onClick={handlePlayToggle}
        disabled={!isLoaded}
        title={
          !isLoaded ? 'Cargando música...' :
          isPlaying ? 'Pausar música' : 'Reproducir música'
        }
      >
        <span className="music-icon">
          {!isLoaded ? '⏳' : isPlaying ? '⏸️' : '🎵'}
        </span>
        <div className={`music-visualizer ${isPlaying && isLoaded ? 'active' : ''}`}>
          <div className="bar bar1"></div>
          <div className="bar bar2"></div>
          <div className="bar bar3"></div>
          <div className="bar bar4"></div>
        </div>
        {!isLoaded && (
          <div className="loading-ring"></div>
        )}
      </button>

      {/* Indicador de estado */}
      <div className={`status-indicator ${loadingState}`}>
        {!isLoaded ? '🔄' : isPlaying ? '🎶' : '⏹️'}
      </div>

      {/* Botón para mostrar/ocultar controles */}
      <button 
        className="music-controls-toggle"
        onClick={() => setShowControls(!showControls)}
        title="Mostrar controles"
        disabled={!isLoaded}
      >
        ⚙️
      </button>

      {/* Panel de controles expandido */}
      {showControls && (
        <div className="music-controls-panel">
          <div className="music-header">
            <span className="music-title">
              🎵 Música Futurista 
              {!isLoaded && <span className="loading-text"> (Cargando...)</span>}
            </span>
            <button 
              className="close-controls"
              onClick={() => setShowControls(false)}
            >
              ❌
            </button>
          </div>

          {/* Información de estado */}
          <div className="status-info">
            <span className={`status-badge ${isLoaded ? 'loaded' : 'loading'}`}>
              {isLoaded ? '✅ Listo' : '⏳ Cargando'}
            </span>
          </div>

          {/* Barra de progreso */}
          <div className="progress-container">
            <span className="time-display">{formatTime(currentTime)}</span>
            <div 
              className={`progress-bar ${!isLoaded ? 'disabled' : ''}`}
              onClick={handleSeek}
            >
              <div 
                className="progress-fill"
                style={{ width: `${progressPercent}%` }}
              ></div>
              <div 
                className="progress-thumb"
                style={{ left: `${progressPercent}%` }}
              ></div>
            </div>
            <span className="time-display">{formatTime(duration)}</span>
          </div>

          {/* Control de volumen */}
          <div className="volume-container">
            <span className="volume-icon">🔊</span>
            <input
              type="range"
              min="0"
              max="1"
              step="0.1"
              value={volume}
              onChange={(e) => changeVolume(parseFloat(e.target.value))}
              className="volume-slider"
              disabled={!isLoaded}
            />
            <span className="volume-display">{Math.round(volume * 100)}%</span>
          </div>

          {/* Botones de control */}
          <div className="playback-controls">
            <button 
              className="control-button"
              onClick={handlePlayToggle}
              disabled={!isLoaded}
            >
              {!isLoaded ? '⏳ Cargando...' : 
               isPlaying ? '⏸️ Pausar' : '▶️ Reproducir'}
            </button>
          </div>

          {/* Debug info (remover en producción) */}
          {process.env.NODE_ENV === 'development' && (
            <div className="debug-info">
              <small>
                Estado: {isLoaded ? 'Cargado' : 'Cargando'} | 
                Reproduciendo: {isPlaying ? 'Sí' : 'No'} | 
                Duración: {formatTime(duration)}
              </small>
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default MusicPlayer
