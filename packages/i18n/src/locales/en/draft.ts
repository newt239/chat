import type { Messages } from "../../messages";

export const draft: Messages["draft"] = {
  list: {
    delete: "Delete draft",
    deleted: "Draft deleted",
    empty: "No drafts",
    emptyHint: "Messages you start writing are saved automatically and listed here",
    inThread: "Thread reply",
    open: "Open",
    savedAt: "Saved {{time}}",
  },
  page: {
    tabs: {
      drafts: "Drafts",
      scheduled: "Scheduled",
      sent: "Sent",
    },
    title: "Drafts & scheduled",
  },
  sidebar: {
    hasDraft: "Draft",
  },
};
