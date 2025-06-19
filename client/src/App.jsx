import { BrowserRouter } from 'react-router-dom'
import { AuthProvider } from './context/AuthContext'
import { MusicProvider } from './context/MusicContext'
import AppRoutes from './routes/AppRoutes'
import MusicPlayer from './components/MusicPlayer'
import './App.css'

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <MusicProvider>
          <MusicPlayer />
          <AppRoutes />
        </MusicProvider>
      </AuthProvider>
    </BrowserRouter>
  )
}

export default App
