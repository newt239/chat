import type { Messages } from "../../messages";

export const common: Messages["common"] = {
  cancel: "Cancel",
  close: "Close",
  copyFailed: "Couldn't copy",
  delete: "Delete",
  loading: "Loading",
  ok: "OK",
  save: "Save",
  upload: {
    aborted: "Upload canceled",
    http: "Upload failed (HTTP {{status}})",
    network: "A network error occurred",
  },
};
