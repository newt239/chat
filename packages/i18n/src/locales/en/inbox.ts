import type { Messages } from "../../messages";

export const inbox: Messages["inbox"] = {
  mention: {
    emptyDescription: "Messages that mention you will show up here",
    emptyTitle: "No mentions",
    failed: "Couldn't load mentions",
    replyPlaceholder: "Reply to {{name}} in thread…",
  },
  replyHint: "Reply right from here",
  thread: {
    open: "Open thread",
    replyPlaceholder: "Reply…",
    showMore: "Show {{count}} more replies",
    unread: "{{count}} unread",
  },
};
