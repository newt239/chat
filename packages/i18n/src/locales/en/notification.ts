import type { Messages } from "../../messages";

export const notification: Messages["notification"] = {
  empty: "No notifications",
  mentionTitle: "Mention from {{name}}",
  remove: "Remove notification",
  type: {
    mention: "Mention",
    message: "Message",
    reaction: "Reaction",
  },
};
