import type { Messages } from "../../messages";

export const common: Messages["common"] = {
  actionFailed: "Something went wrong",
  cancel: "Cancel",
  close: "Close",
  copyFailed: "Couldn't copy",
  delete: "Delete",
  linkCopied: "Link copied",
  retry: "Retry",
  save: "Save",
  upload: {
    http: "Upload failed (HTTP {{status}})",
    network: "A network error occurred",
  },
};
