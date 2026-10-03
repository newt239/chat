import type { Messages } from "../../messages";

export const auth: Messages["auth"] = {
  displayName: "Display name",
  email: "Email",
  google: {
    browserFailed: "Couldn't finish signing in with Google. Please try again.",
    continueInBrowser: "Sign in with Google in your browser",
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
  join: {
    alreadyJoined: "You're already a member",
    lead: "Create an account with Google or with an email and password to join.",
    leadGoogleOnly: "Create an account with Google to join.",
    loggedInLead: "You're logged in as {{name}}.",
    notFound: "The join link was not found or isn't accepting sign-ups",
    open: "Open workspace",
    submit: "Join",
    signUp: "Create account and join",
    title: "Join {{workspace}}",
  },
  login: {
    configFailed: "Couldn't load sign-in options",
    invitationOnly: "Accounts are created by an admin's invitation or from a workspace join link",
    or: "or",
    submit: "Log in",
    title: "Log in",
  },
  logoutConfirm: {
    body: "You will be signed out on this device. You will need to sign in again to continue.",
    title: "Log out?",
  },
  password: "Password",
  passwordRule: "At least 8 characters",
};
