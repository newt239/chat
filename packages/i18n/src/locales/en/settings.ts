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
    desktopDescription: "Show notifications in this browser while the app is open",
    level: "Notify me about",
    levelDescription: "Shared across all your devices",
    levels: {
      all: "All messages",
      mentions: "Mentions and DMs",
      none: "Nothing",
    },
    muteHint:
      "Muted channels and DMs never notify you. Mute from the “More” menu in the channel header.",
    push: "Push notifications",
    pushDescription:
      "Notify this device about mentions, DMs and replies in threads you follow, even when the app is closed",
    pushFailed: "Couldn't turn on push notifications",
  },
  profile: {
    addLink: "Add link",
    avatar: "Avatar",
    bio: "About",
    displayNameDescription: "Shown on your messages and mentions",
    links: "Links",
    linksDescription: "Up to {{max}} URLs. X, Instagram, YouTube, GitHub and more show their icons",
    linkUrl: "Link {{number}}",
    removeLink: "Remove link {{number}}",
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
