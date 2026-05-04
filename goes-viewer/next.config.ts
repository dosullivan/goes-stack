/** @type {import('next').NextConfig} */
const nextConfig = {
  output: 'standalone',
  images: {
    remotePatterns: [
      // Replace this with your own S3 endpoint hostname — must match
      // goes-api-go's S3_ENDPOINT. See README.md ("Image hostname allowlist").
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

