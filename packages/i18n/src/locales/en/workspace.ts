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
  members: {
    invite: "Invite by email",
    inviteSubmit: "Invite",
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
