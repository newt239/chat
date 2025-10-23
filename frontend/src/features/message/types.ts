import type { components } from "@/lib/api/schema";

type BaseMessage = components["schemas"]["Message"];

export interface UserInfo {
  id: string;
  displayName: string;
  avatarUrl?: string | null;
}

export type MessageWithUser = BaseMessage & {
  user: UserInfo;
};

export interface MessagesResponse {
  messages: MessageWithUser[];
  hasMore: boolean;
}
