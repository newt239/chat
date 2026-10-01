import type { ReactNode } from "react";

import { useNavigate, useParams } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
import { Button } from "react-aria-components";

import { cn, focusRing } from "#/components/ui/styles/styles";
import { openPanel } from "#/features/layout/utils/overlaySearch";
import { useMembers } from "#/features/member/hooks/useMembers";
import { useUserGroups } from "#/features/userGroup/hooks/useUserGroups";
import { userAtom } from "#/providers/store/auth";

import { chipClassName } from "./chipClassName";

type MentionProps = {
  "data-mention": string;
  children?: ReactNode;
};

export const Mention = ({ "data-mention": username }: MentionProps) => {
  const { workspaceId } = useParams({ strict: false });
  const { data: members } = useMembers(workspaceId ?? null);
  const { data: groups } = useUserGroups(workspaceId ?? null);
  const currentUser = useAtomValue(userAtom);
  const navigate = useNavigate();

  // バックエンドと同じく、ユーザーは表示名の大文字小文字を区別しない前方一致、グループは名前の完全一致で解決する
  const lowerName = username.toLowerCase();
  const member = members?.find((item) => item.displayName.toLowerCase().startsWith(lowerName));
  const group = member === undefined ? groups?.find((item) => item.name === username) : undefined;

  if (member === undefined && group === undefined) {
    return `@${username}`;
  }

  const isMe = member !== undefined && member.userId === currentUser?.id;
  return (
    <Button
      className={cn(chipClassName, isMe && "bg-mention-chip text-mention-text", focusRing)}
      onPress={() => {
        void navigate({
          search: openPanel(member ? { profile: member.userId } : { group: group?.id }),
          to: ".",
        });
      }}
    >
      @{member?.nickname ?? username}
    </Button>
  );
};
