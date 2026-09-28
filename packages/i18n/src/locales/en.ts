import type { Messages } from "../messages";

export const en: Messages = {
  codeBlock: {
    copied: "Copied",
    copy: "Copy",
    copyFailed: "Couldn't copy",
    lines: "{{count}} lines",
  },
  common: {
    cancel: "Cancel",
    close: "Close",
    delete: "Delete",
    loading: "Loading",
    ok: "OK",
    save: "Save",
  },
  preferences: {
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
  },
  ui: {
    avatar: {
      groupMembers: "Group of {{count}}",
    },
    comboBox: {
      empty: "No suggestions",
      showSuggestions: "Show suggestions",
    },
    toast: {
      dismiss: "Dismiss notification",
      region: "Notifications",
    },
  },
};
