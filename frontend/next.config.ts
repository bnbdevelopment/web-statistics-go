import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  output: 'standalone',
  compress: true,
  poweredByHeader: false,
  experimental: {
    optimizePackageImports: [
      'antd',
      '@ant-design/icons',
      'recharts',
      'reactflow',
      'leaflet',
    ],
  },
};

export default nextConfig;

