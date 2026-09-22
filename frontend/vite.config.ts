import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import path from "node:path";
import { VitePWA } from "vite-plugin-pwa";
import { defineConfig } from "vite-plus";

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
        files: ["tests/**", "*.config.ts"],
        rules: {
          "import/no-nodejs-modules": "off",
          "new-cap": "off",
        },
      },
      {
        files: ["src/lib/logger.ts"],
        rules: {
          "no-console": "off",
        },
      },
      {
        // React Router の Data モードでは redirect() が返す Response を throw する
        files: ["src/routes/routeTree.ts"],
        rules: {
          "typescript/only-throw-error": "off",
        },
      },
      {
        // unified の use() チェーンの型を tsgolint が解決できず error 型になるため
        files: ["src/features/message/utils/markdown/renderer.tsx"],
        rules: {
          "typescript/no-unsafe-assignment": "off",
          "typescript/no-unsafe-call": "off",
          "typescript/no-unsafe-member-access": "off",
          "typescript/no-unsafe-return": "off",
        },
      },
      {
        files: ["**/*.d.ts"],
        rules: {
          "import/unambiguous": "off",
          "typescript/consistent-type-definitions": "off",
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
      // 先頭が大文字化されるとファイルパスの意味が変わるため
      "capitalized-comments": "off",
      // logger のようなインスタンス API を static 化すると設計が歪むため
      "class-methods-use-this": "off",
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
      // 一つ上の階層までの相対インポートは規約で許可している
      "import/no-relative-parent-imports": "off",
      "import/no-unassigned-import": "off",
      "import/prefer-default-export": "off",
      "jsx-a11y/prefer-tag-over-role": "off",
      "max-lines": ["warn", { max: 500, skipBlankLines: true, skipComments: true }],
      "max-lines-per-function": "off",
      "max-statements": "off",
      "new-cap": ["warn", { properties: false }],
      "no-alert": "off",
      // 添付ファイルの逐次アップロードなど順序を保ちたい処理があるため
      "no-await-in-loop": "off",
      "no-console": ["warn", { allow: ["warn", "error"] }],
      // ガード節で continue を使う方がネストが浅く読みやすいため
      "no-continue": "off",
      // 型インポートを別行に分ける方針のため
      "no-duplicate-imports": ["warn", { allowSeparateTypeImports: true }],
      "no-empty-function": "off",
      "no-inline-comments": "off",
      "no-magic-numbers": "off",
      "no-nested-ternary": "off",
      "no-plusplus": "off",
      "no-ternary": "off",
      "no-undefined": "off",
      // Promise を意図的に捨てる `void promise` を許可する
      "no-void": ["warn", { allowAsStatement: true }],
      // TODO コメントは実装予定の記録として残す
      "no-warning-comments": "off",
      "one-var": "off",
      "oxc/no-async-await": "off",
      "oxc/no-barrel-file": "off",
      "oxc/no-optional-chaining": "off",
      "oxc/no-rest-spread-properties": "off",
      "react-perf/jsx-no-jsx-as-prop": "off",
      "react-perf/jsx-no-new-array-as-prop": "off",
      "react-perf/jsx-no-new-function-as-prop": "off",
      "react-perf/jsx-no-new-object-as-prop": "off",
      // 値の変化をトリガーにする effect で本体未参照の依存を意図的に指定しているため
      "react/exhaustive-effect-dependencies": "off",
      "react/forbid-component-props": "off",
      "react/function-component-definition": ["warn", { namedComponents: "arrow-function" }],
      "react/jsx-filename-extension": ["warn", { extensions: [".tsx"] }],
      "react/jsx-max-depth": "off",
      "react/jsx-no-literals": "off",
      "react/jsx-no-useless-fragment": "off",
      "react/jsx-props-no-spreading": "off",
      "react/no-object-type-as-default-prop": "off",
      "react/react-in-jsx-scope": "off",
      // React Compiler が未対応の構文を報告するだけでコード自体は正しいため
      // 外部値とフォーム状態を同期する用途で使っているため
      "react/set-state-in-effect": "off",
      "react/todo": "off",
      "require-unicode-regexp": "off",
      "sort-imports": "off",
      "typescript/consistent-type-definitions": ["error", "type"],
      "typescript/explicit-function-return-type": "off",
      "typescript/explicit-module-boundary-types": "off",
      "typescript/no-empty-interface": "off",
      "typescript/no-non-null-assertion": "error",
      "typescript/no-unsafe-type-assertion": "error",
      // 空文字や false もフォールバックさせたい箇所が多いため
      "typescript/prefer-nullish-coalescing": [
        "warn",
        { ignorePrimitives: { boolean: true, string: true } },
      ],
      "typescript/prefer-readonly-parameter-types": "off",
      // Promise を直接返す関数に async を強制すると require-await と衝突するため
      "typescript/promise-function-async": "off",
      "typescript/strict-boolean-expressions": "off",
      // ファイル名規約は ls-lint 側で PascalCase / camelCase を強制している
      "unicorn/filename-case": "off",
      // zod のスキーマ定義では呼び出しのネストが自然なため
      "unicorn/max-nested-calls": "off",
      "unicorn/no-nested-ternary": "off",
      "unicorn/no-null": "off",
      // consistent-return を満たすための `return undefined` と衝突するため
      "unicorn/no-useless-undefined": "off",
      "unicorn/numeric-separators-style": "off",
      "unicorn/prefer-export-from": "off",
      "unicorn/prefer-global-this": "off",
      "unicorn/prefer-top-level-await": "off",
      // BroadcastChannel.postMessage には targetOrigin がないため誤検出になる
      "unicorn/require-post-message-target-origin": "off",
      "vite-plus/prefer-vite-plus-imports": "error",
    },
  },
  plugins: [
    tailwindcss(),
    react(),
    VitePWA({
      manifest: {
        background_color: "#0b7285",
        display: "standalone",
        lang: "ja",
        name: "Chat",
        short_name: "Chat",
        start_url: "/",
        theme_color: "#0b7285",
      },
      pwaAssets: {
        image: "public/logo.svg",
      },
      registerType: "autoUpdate",
      workbox: {
        runtimeCaching: [
          {
            handler: "NetworkFirst",
            options: {
              cacheName: "api-cache",
              expiration: {
                maxAgeSeconds: 60 * 60 * 24, // 24 hours
                maxEntries: 100,
              },
            },
            urlPattern: /^https:\/\/api\..*/i,
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
    host: true,
    proxy: {
      "/api": {
        changeOrigin: true,
        target: "http://localhost:8080",
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
