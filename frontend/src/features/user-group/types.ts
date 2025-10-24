export type UserGroup = {
  id: string;
  workspaceId: string;
  name: string;
  description?: string;
  createdBy: string;
  createdAt: string;
  updatedAt: string;
};

export type UserGroupMember = {
  groupId: string;
  userId: string;
  joinedAt: string;
};

export type CreateUserGroupInput = {
  workspaceId: string;
  name: string;
  description?: string;
};

export type UpdateUserGroupInput = {
  name?: string;
  description?: string;
};

export type AddMemberInput = {
  email: string;
  role: "admin" | "member";
};
