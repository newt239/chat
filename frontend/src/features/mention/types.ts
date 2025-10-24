export type UserMention = {
  userId: string;
  displayName: string;
};

export type GroupMention = {
  groupId: string;
  name: string;
};

export type MentionSuggestion = {
  id: string;
  name: string;
  type: "user" | "group";
  avatarUrl?: string;
};
