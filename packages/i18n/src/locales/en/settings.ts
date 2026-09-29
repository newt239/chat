import type { Messages } from "../../messages";

export const settings: Messages["settings"] = {
  account: {
    changePassword: "Change password",
    currentPassword: "Current password",
    delete: "Delete account",
    deleteConfirm: "Delete your account?",
    deleteDescription:
      "Your account data, including your messages, will be deleted. This can't be undone.",
    newPassword: "New password",
    openProfile: "Open profile",
    password: "Change password",
    passwordChanged: "Password changed. Please log in again",
    profile: "Profile",
    profileDescription: "Edit your name and bio from your profile",
  },
  notifications: {
    denied: "Notifications are blocked in this browser",
    desktop: "Desktop notifications",
    desktopDescription: "Show notifications in this browser",
    level: "Notify me about",
    levels: {
      all: "All messages",
      mentions: "Mentions and DMs",
      none: "Nothing",
    },
    muteHint:
      "Muted channels and DMs never notify you. Mute from the “More” menu in the channel header.",
  },
  profile: {
    avatarUrl: "Avatar URL",
    bio: "About",
    displayNameDescription: "Shown on your messages and mentions",
    saved: "Profile saved",
  },
  sections: {
    account: "Account",
    display: "Display",
    notifications: "Notifications",
    shortcuts: "Shortcuts",
    theme: "Theme",
  },
  shortcuts: {
    close: "Close panel or dialog",
    newTab: "Open in new tab",
    newline: "New line",
    search: "Search",
    send: "Send",
    settings: "Open settings",
  },
  title: "Settings",
};
