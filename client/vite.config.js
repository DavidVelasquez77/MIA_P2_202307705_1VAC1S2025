import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist'},
  assetsInclude: ['**/*.mp3', '**/*.wav', '**/*.ogg'],
  publicDir: 'public',
  server: {
    fs: {
      allow: ['..']
    }
  }
})
