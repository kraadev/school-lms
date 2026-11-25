/**
 * Generate a self-signed TLS certificate using Node.js v15+ built-in crypto.
 * No external packages needed. Works with Node.js >= 15.6.
 */
const { X509Certificate, createPrivateKey, generateKeyPairSync } = require('crypto')
const fs = require('fs')
const path = require('path')
const os = require('os')
const { execSync } = require('child_process')

const certDir = path.join(__dirname, '..', 'frontend', 'certs')
const keyFile = path.join(certDir, 'key.pem')
const certFile = path.join(certDir, 'cert.pem')

if (!fs.existsSync(certDir)) fs.mkdirSync(certDir, { recursive: true })

// Get local Wi-Fi IP
let localIP = '192.168.43.72'
const ifaces = os.networkInterfaces()
for (const [name, addrs] of Object.entries(ifaces)) {
  for (const addr of addrs) {
    if (addr.family === 'IPv4' && !addr.internal) {
      if (name.toLowerCase().includes('wi-fi') || name.toLowerCase().includes('wlan')) {
        localIP = addr.address
      }
    }
  }
}
console.log(`Target IP: ${localIP}`)

// Use selfsigned package async API correctly
async function generateCert() {
  const selfsignedPath = path.join(__dirname, '..', 'frontend', 'node_modules', 'selfsigned')
  const selfsigned = require(selfsignedPath)
  
  return new Promise((resolve, reject) => {
    selfsigned.generate(
      [{ name: 'commonName', value: localIP }],
      {
        days: 365,
        algorithm: 'sha256',
        keySize: 2048,
        extensions: [
          {
            name: 'subjectAltName',
            altNames: [
              { type: 7, ip: localIP },
              { type: 7, ip: '127.0.0.1' },
              { type: 2, value: 'localhost' },
            ],
          },
        ],
      },
      (err, pems) => {
        if (err) return reject(err)
        resolve(pems)
      }
    )
  })
}

generateCert().then((pems) => {
  console.log('Keys:', Object.keys(pems))
  
  const keyPem = pems.private || pems.serviceKey
  const certPem = pems.cert || pems.certificate

  if (!keyPem || !certPem) {
    console.error('Cannot find key/cert in output. Got:', pems)
    process.exit(1)
  }

  fs.writeFileSync(keyFile, keyPem)
  fs.writeFileSync(certFile, certPem)

  console.log('\n✅ Certificate generated!')
  console.log('  Key :', keyFile)
  console.log('  Cert:', certFile)
  console.log(`\n📱 Access via: https://${localIP}:3000`)
}).catch(err => {
  console.error('Error:', err.message)
  process.exit(1)
})
