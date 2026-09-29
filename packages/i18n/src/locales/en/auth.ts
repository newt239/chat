import type { Messages } from "../../messages";

export const auth: Messages["auth"] = {
  displayName: "Display name",
  email: "Email",
  google: {
    loadFailed: "Couldn't load Google sign-in",
  },
  invite: {
    hasAccount: "Already have an account?",
    lead: "This invitation is for {{email}}. Continue with the Google account for that address.",
    linkTitle: "Invitation link",
    notFound: "The invitation was not found or has expired",
    passwordLead: "If you don't use a Google account, you can join by setting a password.",
    submit: "Set password and join",
    title: "Invitation to {{workspace}}",
  },
  login: {
    invitationOnly: "Accounts are created by invitation from an admin",
    or: "or",
    submit: "Log in",
    title: "Log in",
  },
  password: "Password",
  passwordRule: "At least 8 characters",
};
