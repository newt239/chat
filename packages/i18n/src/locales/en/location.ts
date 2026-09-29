import type { Messages } from "../../messages";

export const location: Messages["location"] = {
  accuracy: "Accuracy about {{meters}} m",
  card: {
    openExternal: "Open in Maps",
    title: "Location",
  },
  composer: {
    remove: "Remove location",
    share: "Share location",
  },
  share: {
    confirm: "Share",
    denied: "Location access is blocked. Check your browser settings.",
    label: "Label (optional)",
    labelPlaceholder: "e.g. Main entrance",
    locating: "Getting your current location…",
    retry: "Try again",
    title: "Share current location",
    unavailable: "Couldn't get your current location",
  },
};
