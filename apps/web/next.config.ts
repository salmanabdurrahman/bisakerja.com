import type { NextConfig } from "next";

import {
  normalizeAPIBasePath,
  resolveAPIOrigin,
} from "./src/lib/config/api-base";

const publicAPIBasePath = normalizeAPIBasePath(
  process.env.NEXT_PUBLIC_API_BASE_URL,
);
const apiOrigin = resolveAPIOrigin(process.env.API_ORIGIN);

const nextConfig: NextConfig = {
  /* config options here */
  reactCompiler: true,
  async rewrites() {
    return [
      {
        source: `${publicAPIBasePath}/:path*`,
        destination: `${apiOrigin}${publicAPIBasePath}/:path*`,
      },
    ];
  },
};

export default nextConfig;
