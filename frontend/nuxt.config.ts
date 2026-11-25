/// <reference types="node" />
import tailwindcss from '@tailwindcss/vite'
import fs from 'node:fs'
import path from 'node:path'
import net from 'node:net'

// Local LAN SSL certificates generated with SAN for 192.168.43.72 and localhost
const BACKEND_URL = process.env.BACKEND_INTERNAL_URL || 'http://localhost:8080'

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  app: {
    head: {
      title: 'School Space',
      meta: [
        { name: 'description', content: 'School Space - Sistem Manajemen Pembelajaran Terpadu' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1.0' },
      ],
      link: [
        { rel: 'icon', type: 'image/png', href: '/images/logo.png' },
      ],
    },
  },
  modules: ['@pinia/nuxt'],
  routeRules: {
    '/api/**': { proxy: `${BACKEND_URL}/api/**` },
    '/uploads/**': { proxy: `${BACKEND_URL}/uploads/**` },
    '/rtc/**': { proxy: 'http://localhost:7880/rtc/**' },
  },
  hooks: {
    'listen'(server: any) {
      const existingListeners = server.rawListeners('upgrade')
      server.removeAllListeners('upgrade')

      server.on('upgrade', (req: any, socket: any, head: any) => {
        if (req.url && (req.url.startsWith('/rtc') || req.url.startsWith('/api/ws'))) {
          const targetPort = req.url.startsWith('/rtc') ? 7880 : 8080
          let targetSocket: any = null
          const safeDestroy = () => {
            try { socket.destroy() } catch {}
            try { targetSocket?.destroy() } catch {}
          }
          socket.on('error', safeDestroy)
          socket.on('close', safeDestroy)

          targetSocket = net.connect(targetPort, '127.0.0.1', () => {
            targetSocket.write(`${req.method} ${req.url} HTTP/${req.httpVersion}\r\n`)
            for (let i = 0; i < req.rawHeaders.length; i += 2) {
              const key = req.rawHeaders[i]
              const val = key.toLowerCase() === 'host' ? `127.0.0.1:${targetPort}` : req.rawHeaders[i + 1]
              targetSocket.write(`${key}: ${val}\r\n`)
            }
            targetSocket.write('\r\n')
            if (head && head.length) {
              targetSocket.write(head)
            }
            targetSocket.pipe(socket)
            socket.pipe(targetSocket)
          })
          targetSocket.on('error', safeDestroy)
          targetSocket.on('close', safeDestroy)
        } else {
          for (const listener of existingListeners) {
            listener.call(server, req, socket, head)
          }
        }
      })
    },
  },
  nitro: {
    devProxy: {
      '/api': {
        target: BACKEND_URL,
        changeOrigin: true,
        ws: true,
      },
      '/uploads': {
        target: BACKEND_URL,
        changeOrigin: true,
      },
      '/rtc': {
        target: 'http://localhost:7880',
        changeOrigin: true,
        ws: true,
      },
    },
  },
  vite: {
    plugins: [tailwindcss()],
    server: {
      host: '0.0.0.0',
      proxy: {
        '/api': {
          target: BACKEND_URL,
          changeOrigin: true,
          secure: false,
          ws: true,
        },
        '/uploads': {
          target: BACKEND_URL,
          changeOrigin: true,
          secure: false,
        },
        '/rtc': {
          target: 'http://localhost:7880',
          changeOrigin: true,
          secure: false,
          ws: true,
        },
      },
      https: (() => {
        const keyPath = path.resolve('./certs/key.pem')
        const certPath = path.resolve('./certs/cert.pem')
        if (fs.existsSync(keyPath) && fs.existsSync(certPath)) {
          return { key: fs.readFileSync(keyPath), cert: fs.readFileSync(certPath) }
        }
        return true
      })(),
    },
  },
  devServer: {
    host: '0.0.0.0',
    https: {
      key: path.resolve('./certs/key.pem'),
      cert: path.resolve('./certs/cert.pem'),
    },
  },
  runtimeConfig: {
    public: {
      // Use relative path — works for both localhost and LAN because requests
      // go through Nuxt's HTTPS server which proxies to the backend.
      apiBase: process.env.NUXT_PUBLIC_API_BASE || '/api',
    },
  },
})

