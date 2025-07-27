/** @type {import('next').NextConfig} */
const nextConfig = {
  env: {
    API_GATEWAY_URL: process.env.API_GATEWAY_URL || 'http://localhost:8080',
    WS_URL: process.env.WS_URL || 'ws://localhost:8084',
  },
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: `${process.env.API_GATEWAY_URL || 'http://localhost:8080'}/api/:path*`,
      },
    ];
  },
  async headers() {
    return [
      {
        source: '/api/:path*',
        headers: [
          { key: 'Access-Control-Allow-Credentials', value: 'true' },
          { key: 'Access-Control-Allow-Origin', value: '*' },
          { key: 'Access-Control-Allow-Methods', value: 'GET,OPTIONS,PATCH,DELETE,POST,PUT' },
          { key: 'Access-Control-Allow-Headers', value: 'X-CSRF-Token, X-Requested-With, Accept, Accept-Version, Content-Length, Content-MD5, Content-Type, Date, X-Api-Version' },
        ],
      },
    ];
  },
  webpack: (config, { buildId, dev, isServer, defaultLoaders, nextRuntime, webpack }) => {
    // Improve dependency resolution for pnpm workspaces
    config.resolve.symlinks = false;
    
    // Add fallbacks for better module resolution
    config.resolve.fallback = {
      ...config.resolve.fallback,
    };

    // Important: return the modified config
    return config;
  },
};

module.exports = nextConfig;