import type { Messages } from "../../messages";

export const command: Messages["command"] = {
  failed: "Couldn't run the command",
  remind: {
    usage: "[me|@user|#channel] what when (e.g. tomorrow 9:00, in 30m)",
  },
};
