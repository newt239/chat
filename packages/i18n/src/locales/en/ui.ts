import type { Messages } from "../../messages";

export const ui: Messages["ui"] = {
  avatar: {
    groupMembers: "Group of {{count}}",
  },
  calendar: {
    next: "Next month",
    previous: "Previous month",
  },
  comboBox: {
    empty: "No suggestions",
    showSuggestions: "Show suggestions",
  },
  copyableUrl: {
    copied: "Copied the link",
    copy: "Copy",
    copyFailed: "Couldn't copy",
  },
  iconImage: {
    change: "Change image",
    reset: "Reset",
    select: "Select image",
    uploadFailed: "Couldn't upload the image. Please try again later",
  },
  imageCrop: {
    apply: "Apply",
    title: "Crop image",
    zoom: "Zoom",
  },
  toast: {
    dismiss: "Dismiss notification",
    region: "Notifications",
  },
};
