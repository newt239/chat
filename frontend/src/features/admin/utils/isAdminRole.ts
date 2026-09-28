import { WorkspaceRole } from "#/gen/chat/v1/workspace_service_pb";

export const isAdminRole = (role: WorkspaceRole | undefined) =>
  role === WorkspaceRole.OWNER || role === WorkspaceRole.ADMIN;
