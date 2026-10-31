import type { NextConfig } from "next";
import createNextIntlPlugin from "next-intl/plugin";

const withNextIntl = createNextIntlPlugin();

const nextConfig: NextConfig = {
  output: "standalone",
  /* config options here */
  webpack(config) {
    config.module.rules.push({
      test: /\.svg$/,
      use: ["@svgr/webpack"],
    });
    return config;
  },
  images: {
    localPatterns: [
      {
        pathname: "/**",
      },
    ],
  },
  turbopack: {
    rules: {
      "*.svg": {
        loaders: ["@svgr/webpack"],
        as: "*.js",
      },
    },
  },
  redirects() {
    return [
      // Alias rute lama (era shadcn): /login dan /register kini /signin dan /signup.
      // Proxy kemudian meneruskan user ber-cookie ke /dashboard.
      {
        source: "/login",
        destination: "/signin",
        permanent: true,
      },
      {
        source: "/register",
        destination: "/signup",
        permanent: true,
      },
      {
        source: "/buttons-group",
        destination: "/buttons",
        permanent: true,
      },
      {
        source: "/modals",
        destination: "/dialogs",
        permanent: true,
      },
      {
        source: "/text-generator",
        destination: "/ai-chat/text",
        permanent: true,
      },
      {
        source: "/image-generator",
        destination: "/ai-chat/image",
        permanent: true,
      },
      {
        source: "/code-generator",
        destination: "/ai-chat/code",
        permanent: true,
      },
      {
        source: "/video-generator",
        destination: "/ai-chat/video",
        permanent: true,
      },
      {
        source: "/ai-settings",
        destination: "/ai-settings/account",
        permanent: true,
      },
    ];
  },
};

export default withNextIntl(nextConfig);
