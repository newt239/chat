import type { Messages } from "../../messages";

export const member: Messages["member"] = {
  note: {
    memo: "Note",
    memoPlaceholder: "Their role, how they like to talk, things to remember",
    nickname: "Display name",
    realName: "Real name: {{name}}",
    saved: "Saved the nickname and note",
    title: "Only visible to you",
  },
  profile: {
    bio: "About",
    email: "Email",
    links: "Links",
    loadFailed: "Couldn't load the profile",
    localTime: "Local time",
    message: "Message",
    notFound: "User not found",
    role: "Role",
  },
  role: {
    admin: "Admin",
    guest: "Guest",
    member: "Member",
    owner: "Owner",
  },
};
