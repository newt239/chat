import { useParams } from "@tanstack/react-router";

import { useMembers } from "./useMembers";

/** 自分が付けたニックネーム、メンバーの表示名、fallback の順に名前を返す関数。ユーザー名の表示はすべてこれを通す */
export const useDisplayName = () => {
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const { data: members } = useMembers(workspaceId);

  return (userId: string, fallback: string) => {
    const member = members?.find((candidate) => candidate.userId === userId);
    return member?.nickname ?? member?.displayName ?? fallback;
  };
};
