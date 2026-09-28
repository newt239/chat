import { defineConfig } from "vite-plus";

export default defineConfig({
  lint: {
    jsPlugins: [{ name: "vite-plus", specifier: "vite-plus/oxlint-plugin" }],
    options: {
      typeAware: true,
      typeCheck: true,
    },
    plugins: ["eslint", "typescript", "unicorn", "oxc", "import"],
    rules: {
      "func-style": ["error", "expression"],
      "typescript/consistent-type-definitions": ["error", "type"],
      "typescript/no-non-null-assertion": "error",
      "typescript/no-unsafe-type-assertion": "error",
      "vite-plus/prefer-vite-plus-imports": "error",
    },
  },
  test: {
    include: ["src/**/*.spec.ts"],
  },
});
