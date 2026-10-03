import type { Messages } from "../../messages";

export const inbox: Messages["inbox"] = {
  mention: {
    emptyDescription: "Messages that mention you will show up here",
    emptyTitle: "No mentions",
    failed: "Couldn't load mentions",
    replyPlaceholder: "Reply to {{name}} in thread…",
  },
};
