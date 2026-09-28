import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterEach, vi } from "vite-plus/test";

import "#/lib/i18n";

// jsdom は matchMedia を実装していないため、どのメディアクエリにも一致しない実装で置き換える
vi.stubGlobal("matchMedia", (query: string) => ({
  addEventListener: () => {},
  matches: false,
  media: query,
  removeEventListener: () => {},
}));

// Node 26 の組み込み localStorage が jsdom のものを覆い隠し、未設定だと undefined になるため
const memoryStorage = new Map<string, string>();
vi.stubGlobal("localStorage", {
  clear: () => {
    memoryStorage.clear();
  },
  getItem: (key: string) => memoryStorage.get(key) ?? null,
  key: (index: number) => [...memoryStorage.keys()][index] ?? null,
  get length() {
    return memoryStorage.size;
  },
  removeItem: (key: string) => {
    memoryStorage.delete(key);
  },
  setItem: (key: string, value: string) => {
    memoryStorage.set(key, value);
  },
});

// globals: false では Testing Library の自動クリーンアップが登録されないため
afterEach(() => {
  cleanup();
});
