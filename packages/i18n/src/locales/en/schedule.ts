import type { Messages } from "../../messages";

export const schedule: Messages["schedule"] = {
  dialog: {
    confirm: "Schedule",
    field: "Send at",
    past: "Pick a time in the future",
    title: "Schedule message",
  },
  failed: "Couldn't schedule the message",
  menu: {
    custom: "Custom time…",
    label: "Schedule send",
  },
  presets: {
    inOneHour: "In 1 hour",
    nextMonday: "Next Monday 9:00",
    tomorrowMorning: "Tomorrow 9:00",
  },
  scheduled: "Scheduled for {{time}}",
};
