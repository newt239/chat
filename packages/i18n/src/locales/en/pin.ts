import type { Messages } from "../../messages";

export const pin: Messages["pin"] = {
  empty: "No pinned messages",
  emptyHint: "Pin important messages so everyone in the channel can find them here",
  label: "Pinned by {{name}}",
  loadFailed: "Couldn't load pinned messages",
  pinned: "Pinned",
  unpinned: "Unpinned",
};
