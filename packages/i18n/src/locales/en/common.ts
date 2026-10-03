import type { Messages } from "../../messages";

export const common: Messages["common"] = {
  cancel: "Cancel",
  close: "Close",
  copyFailed: "Couldn't copy",
  delete: "Delete",
  linkCopied: "Link copied",
  save: "Save",
  upload: {
    aborted: "Upload canceled",
    http: "Upload failed (HTTP {{status}})",
    network: "A network error occurred",
  },
};
