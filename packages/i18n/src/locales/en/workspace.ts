import type { Messages } from "../../messages";

export const workspace: Messages["workspace"] = {
  create: {
    description: "Description (optional)",
    id: "Workspace ID",
    idDescription: "Used in the URL. 3–12 lowercase letters, numbers, and hyphens",
    idInvalid: "Use 3–12 lowercase letters, numbers, and hyphens (not at the start or end)",
    name: "Workspace name",
    submit: "Create",
    title: "Create workspace",
  },
  list: {
    join: "Join",
    loadFailed: "Couldn't load workspaces",
    memberCount: "Members: {{count}}",
    open: "Open",
    public: "Public workspaces you can join",
    title: "Workspaces",
  },
  invite: {
    addedDirectly: "Added {{email}} to the workspace",
    copied: "Copied the invitation link",
    copy: "Copy",
    copyFailed: "Couldn't copy",
    email: "Invite by email",
    failed: "Couldn't invite",
    link: "Invitation link for {{email}}",
    linkOnce:
      "This link is shown only now. Copy it and share it with the person you invited (valid for 7 days).",
    role: "Role to invite as",
    submit: "Invite",
  },
  members: {
    remove: "Remove {{name}} from the workspace",
    role: "Role",
    title: "Members ({{count}})",
  },
  settings: {
    delete: "Delete workspace",
    deleteConfirm: "Delete {{name}}?",
    deleteDescription: "All channels and messages are deleted too. This can't be undone.",
    description: "Description",
    isPublic: "Make this workspace public (anyone can join)",
    name: "Name",
    title: "Workspace settings",
  },
};
