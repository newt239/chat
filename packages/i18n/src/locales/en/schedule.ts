import type { Messages } from "../../messages";

export const schedule: Messages["schedule"] = {
  dialog: {
    confirm: "Schedule",
    field: "Send at",
    past: "Pick a time in the future",
    title: "Schedule message",
  },
  failed: "Couldn't schedule the message",
  list: {
    actionFailed: "Something went wrong",
    actions: "Scheduled message actions",
    attachments: "{{count}} attachments",
    body: "Message",
    deleted: "Scheduled message deleted",
    edit: "Edit",
    editTitle: "Edit scheduled message",
    empty: "No scheduled messages",
    emptyHint: "Use the menu next to the send button to schedule a message",
    scheduledFor: "Sends {{time}}",
    sendNow: "Send now",
    sent: "Sent",
    sentAt: "Sent {{time}}",
    sentEmpty: "No scheduled messages have been sent yet",
    sentEmptyHint: "Scheduled messages are sent automatically and show up here",
    showMessage: "View message",
    status: {
      failed: "Failed",
      sending: "Sending",
    },
    updated: "Scheduled message updated",
  },
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
