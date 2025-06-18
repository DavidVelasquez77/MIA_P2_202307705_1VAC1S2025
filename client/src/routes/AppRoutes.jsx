import { Routes, Route, Navigate } from 'react-router-dom'
import Login from '../components/Login'
import Console from '../components/Console'
import FileSystemViewer from '../components/FileSystemViewer'
import ProtectedRoute from './ProtectedRoute'

function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<Console />} />
      <Route path="/console" element={<Console />} />
      <Route path="/login" element={<Login />} />
      <Route 
        path="/filesystem" 
        element={
          <ProtectedRoute>
            <FileSystemViewer />
          </ProtectedRoute>
        } 
      />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default AppRoutes
