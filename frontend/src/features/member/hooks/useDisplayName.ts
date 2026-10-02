import { useCallback } from "react";

import { useParams } from "@tanstack/react-router";

import { useMembers } from "./useMembers";

/** 自分が付けたニックネームがあればそれを、なければ表示名を返す関数。ユーザー名の表示はすべてこれを通す */
export const useDisplayName = () => {
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const { data: members } = useMembers(workspaceId);

  return useCallback(
    (userId: string, displayName: string) =>
      members?.find((member) => member.userId === userId)?.nickname ?? displayName,
    [members],
  );
};
