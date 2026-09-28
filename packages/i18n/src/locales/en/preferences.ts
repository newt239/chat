import type { Messages } from "../../messages";

export const preferences: Messages["preferences"] = {
  locale: {
    en: "English",
    ja: "日本語",
    title: "Language",
  },
  mode: {
    dark: "Dark",
    light: "Light",
    system: "Match system",
    title: "Appearance",
  },
  saveFailed: "Couldn't save your preferences",
  theme: {
    chroma: "Saturation",
    custom: "Custom",
    hue: "Hue",
    presets: {
      cobalt: "Cobalt",
      graphite: "Graphite",
      jade: "Jade",
      plum: "Plum",
    },
    sidebar: {
      light: "Light",
      tinted: "Tinted",
      title: "Sidebar",
    },
    title: "Theme",
  },
};
