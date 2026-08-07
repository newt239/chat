import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import path from "node:path";
import { VitePWA } from "vite-plugin-pwa";
import { defineConfig } from "vite-plus";

// 自動生成ファイルは lint / format の対象外にする
const generatedFiles = ["src/lib/api/schema.ts"];

export default defineConfig({
  fmt: {
    ignorePatterns: ["dist/", "dev-dist/", ...generatedFiles],
    jsdoc: true,
    sortImports: {
      customGroups: [
        {
          elementNamePattern: ["react", "react-dom"],
          groupName: "react",
        },
      ],
      groups: [
        "react",
        ["value-builtin", "value-external"],
        "value-internal",
        ["value-parent", "value-sibling", "value-index"],
        ["type-parent", "type-sibling", "type-index"],
        "type-internal",
        "type-import",
        "unknown",
      ],
      ignoreCase: true,
      internalPattern: ["#/"],
      newlinesBetween: true,
      order: "asc",
    },
    sortPackageJson: {
      sortScripts: true,
    },
  },
  lint: {
    categories: {
      nursery: "warn",
      pedantic: "warn",
      perf: "warn",
      restriction: "warn",
      style: "warn",
      suspicious: "warn",
    },
    env: {
      browser: true,
      node: true,
    },
    ignorePatterns: ["dist/", "dev-dist/", ...generatedFiles],
    jsPlugins: [{ name: "vite-plus", specifier: "vite-plus/oxlint-plugin" }],
    options: {
      typeAware: true,
      typeCheck: true,
    },
    overrides: [
      {
        files: ["*.config.ts"],
        rules: {
          "import/no-nodejs-modules": "off",
        },
      },
      {
        files: ["**/*.d.ts"],
        rules: {
          "import/unambiguous": "off",
        },
      },
    ],
    plugins: [
      "eslint",
      "typescript",
      "unicorn",
      "react",
      "react-perf",
      "oxc",
      "import",
      "jsx-a11y",
    ],
    rules: {
      complexity: "off",
      "func-style": ["error", "expression"],
      "id-length": "off",
      "import/consistent-type-specifier-style": "off",
      "import/exports-last": "off",
      "import/group-exports": "off",
      "import/max-dependencies": "off",
      "import/no-default-export": "off",
      "import/no-named-export": "off",
      "import/no-namespace": "off",
      "import/no-unassigned-import": "off",
      "import/prefer-default-export": "off",
      "max-lines": ["warn", { max: 500, skipBlankLines: true, skipComments: true }],
      "max-lines-per-function": "off",
      "max-statements": "off",
      "no-alert": "off",
      "no-console": ["warn", { allow: ["warn", "error"] }],
      "no-inline-comments": "off",
      "no-magic-numbers": "off",
      "no-nested-ternary": "off",
      "no-plusplus": "off",
      "no-ternary": "off",
      "no-undefined": "off",
      "oxc/no-async-await": "off",
      "oxc/no-barrel-file": "off",
      "oxc/no-optional-chaining": "off",
      "oxc/no-rest-spread-properties": "off",
      "react-perf/jsx-no-new-array-as-prop": "off",
      "react-perf/jsx-no-new-function-as-prop": "off",
      "react-perf/jsx-no-new-object-as-prop": "off",
      "react/function-component-definition": ["warn", { namedComponents: "arrow-function" }],
      "react/jsx-filename-extension": ["warn", { extensions: [".tsx"] }],
      "react/jsx-max-depth": "off",
      "react/jsx-no-literals": "off",
      "react/jsx-no-useless-fragment": "off",
      "react/jsx-props-no-spreading": "off",
      "react/react-in-jsx-scope": "off",
      "sort-imports": "off",
      // 既存コード 8,600 行に対して数百件の警告が出て出力が読めなくなるため無効化する
      "sort-keys": "off",
      "typescript/consistent-type-definitions": ["error", "type"],
      "typescript/explicit-function-return-type": "off",
      "typescript/explicit-module-boundary-types": "off",
      "typescript/no-empty-interface": "off",
      "typescript/no-non-null-assertion": "off",
      "typescript/no-unsafe-type-assertion": "off",
      "typescript/prefer-readonly-parameter-types": "off",
      "typescript/strict-boolean-expressions": "off",
      "unicorn/no-nested-ternary": "off",
      "unicorn/no-null": "off",
      "unicorn/numeric-separators-style": "off",
      "unicorn/prefer-global-this": "off",
      "vite-plus/prefer-vite-plus-imports": "error",
    },
  },
  plugins: [
    tailwindcss(),
    react(),
    VitePWA({
      registerType: "autoUpdate",
      // public/logo.svg から各サイズのアイコンを生成し manifest の icons に注入する
      pwaAssets: {
        image: "public/logo.svg",
      },
      manifest: {
        name: "Chat",
        short_name: "Chat",
        lang: "ja",
        start_url: "/",
        display: "standalone",
        background_color: "#0b7285",
        theme_color: "#0b7285",
      },
      workbox: {
        runtimeCaching: [
          {
            urlPattern: /^https:\/\/api\..*/i,
            handler: "NetworkFirst",
            options: {
              cacheName: "api-cache",
              expiration: {
                maxEntries: 100,
                maxAgeSeconds: 60 * 60 * 24, // 24 hours
              },
            },
          },
        ],
      },
    }),
  ],
  resolve: {
    alias: {
      "#": path.resolve(import.meta.dirname, "src"),
    },
  },
  server: {
    // Docker コンテナ外からアクセスできるようにする
    host: true,
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
      "/ws": {
        target: "ws://localhost:8080",
        ws: true,
      },
    },
  },
  test: {
    environment: "jsdom",
    globals: false,
    include: ["src/**/*.{spec,test}.{ts,tsx}"],
    setupFiles: ["./tests/vitest.setup.ts"],
  },
});
