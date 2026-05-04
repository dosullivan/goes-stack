/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  images: {
    remotePatterns: [
      {
        protocol: 'http',
        hostname: 'rustfs.int.ridge.casa',
        port: '9000',
        pathname: '/**',
      },
    ],
  },
}

module.exports = nextConfig

