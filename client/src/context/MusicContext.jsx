import { createContext, useContext, useState, useRef, useEffect } from 'react'

const MusicContext = createContext()

export function useMusic() {
  const context = useContext(MusicContext)
  if (!context) {
    throw new Error('useMusic debe ser usado dentro de MusicProvider')
  }
  return context
}

export function MusicProvider({ children }) { 
  const [isPlaying, setIsPlaying] = useState(false)
  const [volume, setVolume] = useState(0.3)
  const [currentTime, setCurrentTime] = useState(0)
  const [duration, setDuration] = useState(0)
  const [isLoaded, setIsLoaded] = useState(false)
  const audioRef = useRef(null)

  useEffect(() => {
    const audio = new Audio()
    audioRef.current = audio
    
    // Intentar múltiples rutas para el archivo
    const audioSources = [
      '/music/309.mp3',
      '/src/music/309.mp3',
      './music/309.mp3',
      './src/music/309.mp3'
    ]

    const loadAudio = async () => {
      for (const src of audioSources) {
        try {
          audio.src = src
          await new Promise((resolve, reject) => {
            audio.addEventListener('canplaythrough', resolve, { once: true })
            audio.addEventListener('error', reject, { once: true })
            audio.load()
          })
          console.log(`Audio cargado desde: ${src}`)
          setIsLoaded(true)
          break
        } catch (error) {
          console.warn(`No se pudo cargar audio desde: ${src}`)
        }
      }
      
      if (!isLoaded) {
        console.error('No se pudo cargar el archivo de audio desde ninguna ruta')
      }
    }

    const updateTime = () => setCurrentTime(audio.currentTime)
    const updateDuration = () => setDuration(audio.duration)
    const handleEnded = () => {
      setIsPlaying(false)
      audio.currentTime = 0
      // Auto-replay para música continua
      setTimeout(() => {
        audio.play().then(() => setIsPlaying(true)).catch(console.error)
      }, 1000)
    }
    const handleLoadedMetadata = () => {
      setDuration(audio.duration)
      setIsLoaded(true)
    }
    const handleError = (e) => {
      console.error('Error de audio:', e)
      setIsLoaded(false)
    }

    // Event listeners
    audio.addEventListener('timeupdate', updateTime)
    audio.addEventListener('loadedmetadata', handleLoadedMetadata)
    audio.addEventListener('ended', handleEnded)
    audio.addEventListener('error', handleError)
    audio.addEventListener('canplaythrough', () => setIsLoaded(true))

    // Configuración inicial
    audio.volume = volume
    audio.loop = false // Manejamos el loop manualmente para mejor control

    // Cargar audio
    loadAudio()

    return () => {
      audio.removeEventListener('timeupdate', updateTime)
      audio.removeEventListener('loadedmetadata', handleLoadedMetadata)
      audio.removeEventListener('ended', handleEnded)
      audio.removeEventListener('error', handleError)
      audio.removeEventListener('canplaythrough', () => setIsLoaded(true))
      audio.pause()
      audio.src = ''
    }
  }, [])

  useEffect(() => {
    if (audioRef.current) {
      audioRef.current.volume = volume
    }
  }, [volume])

  const togglePlay = async () => {
    const audio = audioRef.current
    if (!audio || !isLoaded) {
      console.warn('Audio no está cargado aún')
      return
    }

    try {
      if (isPlaying) {
        audio.pause()
        setIsPlaying(false)
      } else {
        // Asegurar que el audio esté listo
        if (audio.readyState >= 2) {
          await audio.play()
          setIsPlaying(true)
        } else {
          console.warn('Audio no está listo para reproducir')
        }
      }
    } catch (error) {
      console.error('Error al reproducir audio:', error)
      setIsPlaying(false)
    }
  }

  const changeVolume = (newVolume) => {
    setVolume(newVolume)
    if (audioRef.current) {
      audioRef.current.volume = newVolume
    }
  }

  const seek = (time) => {
    if (audioRef.current && isLoaded) {
      audioRef.current.currentTime = time
      setCurrentTime(time)
    }
  }

  const value = {
    isPlaying,
    volume,
    currentTime,
    duration,
    isLoaded,
    togglePlay,
    changeVolume,
    seek
  }

  return (
    <MusicContext.Provider value={value}>
      {children}
    </MusicContext.Provider>
  )
}
