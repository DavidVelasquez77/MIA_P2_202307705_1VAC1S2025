import { Routes, Route, Navigate } from 'react-router-dom'
import Login from '../components/Login'
import Console from '../components/Console'

function AppRoutes() {
  return (
    <Routes>
      <Route path="/" element={<Console />} />
      <Route path="/console" element={<Console />} />
      <Route path="/login" element={<Login />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default AppRoutes
