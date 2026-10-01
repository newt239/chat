import type { Messages } from "../../messages";

export const preferences: Messages["preferences"] = {
  channelSort: {
    description:
      "How channels are ordered in the sidebar. Recent activity lists nested channels separately",
    title: "Channel order",
  },
  joinMessages: {
    description: "Show a notice in the timeline when someone joins a channel",
    title: "Show join messages",
  },
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
    customActive: "In use",
    hue: "Hue",
    presetsTitle: "Presets",
    presets: {
      cobalt: "Cobalt",
      graphite: "Graphite",
      jade: "Jade",
      plum: "Plum",
    },
    preview: "Preview",
    sidebar: {
      light: "Light",
      tinted: "Tinted",
      title: "Sidebar",
    },
    title: "Theme",
  },
  timezone: {
    autoUpdate: "Update time zone automatically",
    autoUpdateDescription: "Update without asking when this device's time zone changes",
    changed: "This device is set to {{timezone}}",
    changedDescription: "Update your account's time zone ({{current}})?",
    description:
      "Shown as your local time on your profile, and used to display dates and read /remind times",
    placeholder: "Not set",
    title: "Time zone",
    update: "Update",
  },
};
