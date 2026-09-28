import type { Messages } from "../../messages";

export const member: Messages["member"] = {
  note: {
    memo: "Note",
    memoPlaceholder: "Their role, how they like to talk, things to remember",
    nickname: "Display name",
    private: "Only you can see this. They won't be notified",
    realName: "Real name: {{name}}",
    saved: "Saved the nickname and note",
    title: "Only visible to you",
  },
  profile: {
    bio: "About",
    email: "Email",
    loadFailed: "Couldn't load the profile",
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
