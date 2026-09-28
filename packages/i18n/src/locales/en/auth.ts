import type { Messages } from "../../messages";

export const auth: Messages["auth"] = {
  displayName: "Display name",
  email: "Email",
  login: {
    noAccount: "Don't have an account?",
    submit: "Log in",
    title: "Log in",
  },
  password: "Password",
  passwordRule: "At least 8 characters",
  register: {
    hasAccount: "Already have an account?",
    submit: "Sign up",
    title: "Sign up",
  },
};
