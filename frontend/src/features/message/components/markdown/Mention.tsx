import type { ReactNode } from "react";

import { useParams } from "@tanstack/react-router";
import { useAtomValue, useSetAtom } from "jotai";
import { Button } from "react-aria-components";

import { cn, focusRing } from "#/components/ui/styles";
import { useMembers } from "#/features/member/hooks/useMembers";
import { userAtom } from "#/providers/store/auth";
import { setRightSidePanelViewAtom } from "#/providers/store/ui";

import { chipClassName } from "./chipClassName";

type MentionProps = {
  "data-mention": string;
  children?: ReactNode;
};

export const Mention = ({ "data-mention": username }: MentionProps) => {
  const { workspaceId } = useParams({ strict: false });
  const { data: members } = useMembers(workspaceId ?? null);
  const currentUser = useAtomValue(userAtom);
  const setRightSidePanelView = useSetAtom(setRightSidePanelViewAtom);

  // メンションは表示名の前方一致で解決される
  const member = members?.find((item) => item.displayName.startsWith(username));
  const isMe = member !== undefined && member.userId === currentUser?.id;
  const className = cn(chipClassName, isMe && "bg-mention-chip text-mention-text");

  if (member === undefined) {
    return <span className={className}>@{username}</span>;
  }

  return (
    <Button
      className={cn(className, focusRing)}
      onPress={() => {
        setRightSidePanelView({ type: "user-profile", userId: member.userId });
      }}
    >
      @{member.nickname ?? username}
    </Button>
  );
};
