import { useCallback } from "react";

import { useAtomValue } from "jotai";

import { currentWorkspaceIdAtom } from "#/providers/store/workspace";

import { useMembers } from "./useMembers";

/** 自分が付けたニックネームがあればそれを、なければ表示名を返す関数。ユーザー名の表示はすべてこれを通す */
export const useDisplayName = () => {
  const workspaceId = useAtomValue(currentWorkspaceIdAtom);
  const { data: members } = useMembers(workspaceId);

  return useCallback(
    (userId: string, displayName: string) =>
      members?.find((member) => member.userId === userId)?.nickname ?? displayName,
    [members],
  );
};
