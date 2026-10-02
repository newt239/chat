import type { ReactNode } from "react";

import { useNavigate } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { cn, focusRing } from "#/components/ui/styles/styles";
import { openPanel } from "#/features/layout/utils/overlaySearch";
import { userAtom } from "#/providers/store/auth";

import { useMentionDirectory } from "../../hooks/useMentionDirectory";
import { chipClassName } from "./chipClassName";

type MentionProps = {
  // 「user:ID」「group:ID」「broadcast:channel」の形
  "data-mention": string;
  children?: ReactNode;
};

/** 本文に ID で埋め込んだメンションを今の名前で出す。名前が変わってもメンション先は変わらない */
export const Mention = ({ "data-mention": value }: MentionProps) => {
  const { t } = useTranslation();
  const directory = useMentionDirectory();
  const currentUser = useAtomValue(userAtom);
  const navigate = useNavigate();
  const [kind, id = ""] = value.split(":");

  if (kind === "broadcast") {
    return (
      <span className={cn(chipClassName, "cursor-default bg-mention-chip text-mention-text")}>
        @{id}
      </span>
    );
  }

  const member = kind === "user" ? directory.member(id) : undefined;
  const group = kind === "group" ? directory.group(id) : undefined;
  if (member === undefined && group === undefined) {
    return (
      <span className={cn(chipClassName, "cursor-default")}>
        @{t(kind === "group" ? "message.mention.unknownGroup" : "message.mention.unknownUser")}
      </span>
    );
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
      @{member ? (member.nickname ?? member.displayName) : group?.name}
    </Button>
  );
};
