import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(({ mode }) => {
  const isCapacitor = mode === 'capacitor'

  return {
    plugins: [vue()],
    base: isCapacitor ? '/' : '/messenger/',
    build: {
      rollupOptions: {
        // Capacitor plugins are only used via dynamic import() in native mode.
        // Mark them as external for web builds so Rollup doesn't try to resolve them.
        external: isCapacitor ? [] : [
          '@capacitor/push-notifications',
          '@capacitor/device',
          '@capacitor/app'
        ]
      }
    },
    server: {
      port: 5173,
      proxy: {
        '/messenger/api': {
          target: 'http://localhost:8080',
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/messenger/, '')
        },
        '/messenger/ws': {
          target: 'ws://localhost:8080',
          ws: true,
          rewrite: (path) => path.replace(/^\/messenger/, '')
        }
      }
    }
  }
})
