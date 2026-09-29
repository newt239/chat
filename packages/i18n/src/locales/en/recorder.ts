import type { Messages } from "../../messages";

export const recorder: Messages["recorder"] = {
  attach: "Attach",
  discard: "Discard recording",
  failed: {
    denied: "Microphone access is blocked. Check your browser settings.",
    unsupported: "This browser can't record audio",
  },
  preparing: "Preparing the microphone…",
  recording: "Recording",
  start: "Record audio",
  stop: "Stop",
};
