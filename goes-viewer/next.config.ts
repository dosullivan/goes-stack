/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  images: {
    remotePatterns: [
      {
        protocol: 'http',
        hostname: 'minio.int.ridge.casa',
        port: '9000',
        pathname: '/**',
      },
    ],
  },
}

module.exports = nextConfig

