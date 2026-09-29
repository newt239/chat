import type { ReactNode } from "react";

import { useNavigate, useParams } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
import { Button } from "react-aria-components";

import { cn, focusRing } from "#/components/ui/styles/styles";
import { openPanel } from "#/features/layout/utils/overlaySearch";
import { useMembers } from "#/features/member/hooks/useMembers";
import { userAtom } from "#/providers/store/auth";

import { chipClassName } from "./chipClassName";

type MentionProps = {
  "data-mention": string;
  children?: ReactNode;
};

export const Mention = ({ "data-mention": username }: MentionProps) => {
  const { workspaceId } = useParams({ strict: false });
  const { data: members } = useMembers(workspaceId ?? null);
  const currentUser = useAtomValue(userAtom);
  const navigate = useNavigate();

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
        void navigate({ search: openPanel({ profile: member.userId }), to: "." });
      }}
    >
      @{member.nickname ?? username}
    </Button>
  );
};
