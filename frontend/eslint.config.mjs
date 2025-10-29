import eslintConfigPrettier from "eslint-config-prettier";
import importPlugin from "eslint-plugin-import";
import react from "eslint-plugin-react";
import reactHooks from "eslint-plugin-react-hooks";
import reactRefresh from "eslint-plugin-react-refresh";
import tseslint from "typescript-eslint";

const eslintConfig = [
  {
    ignores: ["node_modules/", "dist/", "storybook-static/", ".storybook/", "routeTree.gen.ts", "lib/api/schema.ts"],
  },
  ...tseslint.configs.recommended,
  {
    files: ["**/*.{ts,tsx}"],
    settings: {
      react: {
        version: "detect",
      },
    },
    plugins: {
      react,
      "react-hooks": reactHooks,
      "react-refresh": reactRefresh,
      import: importPlugin,
    },
    rules: {
      ...react.configs.recommended.rules,
      ...reactHooks.configs.recommended.rules,
      "no-restricted-imports": [
        "error",
        {
          "patterns": [{
            group: ["../../*"],
            message: '2階層より上のファイルをインポートする場合は@エイリアスを使用した絶対パスで記述してください。',
          },]
        }
      ],
      "@typescript-eslint/no-explicit-any": "error",
      "@typescript-eslint/consistent-type-definitions": ["error", "type"],
      "no-unused-vars": "off", // typescript-eslint の no-unused-vars と競合するため無効化
      "@typescript-eslint/no-unused-vars": "error",
      "@typescript-eslint/consistent-type-imports": [
        "error",
        {
          prefer: "type-imports",
          fixStyle: "separate-type-imports",
          disallowTypeAnnotations: false,
        },
      ],
      "react/react-in-jsx-scope": "off",
      "react-hooks/exhaustive-deps": "off",
      "react/function-component-definition": [
        "error",
        {
          namedComponents: "arrow-function",
          unnamedComponents: "arrow-function",
        },
      ],
      "react/no-multi-comp": "error",
      "import/order": [
        "error",
        {
          groups: [
            "builtin",
            "external",
            "parent",
            "sibling",
            "index",
            "object",
            "type",
          ],
          pathGroups: [
            {
              pattern: "{react,react-dom/**,react-router-dom}",
              group: "builtin",
              position: "before",
            },
          ],
          pathGroupsExcludedImportTypes: ["builtin"],
          alphabetize: {
            order: "asc",
          },
          "newlines-between": "always",
        },
      ],
    },
  },
  {
    files: ["**/*.tsx"],
    rules: {
      "max-lines": ["error", {
        "max": 500,
        "skipBlankLines": true,
        "skipComments": true
      }],
    }
  },
  {
    files: ["**/*.test.{ts,tsx}"],
    rules: {
      "react/no-multi-comp": "off",
    }
  },
  eslintConfigPrettier,
];

export default eslintConfig;
