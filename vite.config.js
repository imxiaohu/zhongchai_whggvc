import uni from '@dcloudio/vite-plugin-uni'
import path from 'path'
import fs from 'fs'

function localApiPlugin() {
  return {
    name: 'local-api-server',
    configureServer(server) {
      server.middlewares.use((req, res, next) => {
        const url = req.url || ''
        const match = url.match(/^\/api\/(.+)\.js(\?.*)?$/)
        if (match) {
          const filePath = path.resolve(__dirname, 'api', match[1] + '.js')
          if (fs.existsSync(filePath)) {
            const content = fs.readFileSync(filePath, 'utf-8')
            res.setHeader('Content-Type', 'application/javascript')
            res.setHeader('Access-Control-Allow-Origin', '*')
            res.end(content)
            return
          }
        }
        next()
      })
    }
  }
}

export default {
  plugins: [uni(), localApiPlugin()],
  server: {
    port: 5174,
    proxy: {
      '/api': {
        target: 'http://localhost:2333',
        changeOrigin: true,
        secure: false,
        ws: false,
        bypass: (req, res) => {
          const url = req.url || ''
          if (/^\/api\/[^?]*\.(js|ts|mjs)(\?.*)?$/.test(url)) {
            return false
          }
        }
      },
      '/scloud': {
        target: 'http://localhost:2333',
        changeOrigin: true,
        secure: false,
        ws: false
      },
      '/scloudoa': {
        target: 'http://localhost:2333',
        changeOrigin: true,
        secure: false,
        ws: false
      }
    }
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, '.'),
      'api': path.resolve(__dirname, './api'),
      'pages/api': path.resolve(__dirname, './pages/api')
    }
  }
}
