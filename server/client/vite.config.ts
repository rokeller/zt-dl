import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
    plugins: [react()],
    server: {
        proxy: {
            '/api/queue/events': {
                target: 'http://localhost:8080',
                ws: true,
            },
            '/api': 'http://localhost:8080',
        },
    },
})
