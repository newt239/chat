import type { Messages } from "../../messages";

export const member: Messages["member"] = {
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
