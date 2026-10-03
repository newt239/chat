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
    copy: "Copy",
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
  menu: {
    title: "Menu",
  },
  pagination: {
    next: "Next page",
    page: "Page {{page}} of {{total}}",
    previous: "Previous page",
  },
  toast: {
    dismiss: "Dismiss notification",
    region: "Notifications",
  },
};
