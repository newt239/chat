import { useNavigate } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
import { Button } from "react-aria-components";

import { Link } from "#/components/ui/Link/Link";
import { cn, focusRing } from "#/components/ui/styles/styles";
import { openPanel } from "#/lib/overlaySearch";
import { myUserIdAtom } from "#/providers/store/auth";

import { useMentionDirectory } from "../hooks/useMentionDirectory";

import type { MentionToken } from "../utils/mentionToken";

const chipClassName =
  "inline cursor-pointer rounded-sm bg-accent-soft px-0.75 font-semibold text-accent-text no-underline";

type MentionProps = {
  token: MentionToken;
};

/** 本文に ID で埋め込んだメンションやチャンネルを今の名前で出す。名前が変わっても指す先は変わらない */
export const Mention = ({ token }: MentionProps) => {
  const directory = useMentionDirectory();
  const myId = useAtomValue(myUserIdAtom);
  const navigate = useNavigate();
  const label = directory.textOf(token);
  const channel = token.kind === "channel" ? directory.channel(token.id) : undefined;
  const member = token.kind === "user" ? directory.member(token.id) : undefined;
  const group = token.kind === "group" ? directory.group(token.id) : undefined;

  if (channel && directory.workspaceId !== null) {
    return (
      <Link
        to="/app/$workspaceId/$channelId"
        params={{ channelId: channel.id, workspaceId: directory.workspaceId }}
        className={chipClassName}
      >
        {label}
      </Link>
    );
  }
  if (member === undefined && group === undefined) {
    return (
      <span
        className={cn(
          chipClassName,
          "cursor-default",
          token.kind === "broadcast" && "bg-mention-chip text-mention-text",
        )}
      >
        {label}
      </span>
    );
  }
  return (
    <Button
      className={cn(
        chipClassName,
        member?.userId === myId && "bg-mention-chip text-mention-text",
        focusRing,
      )}
      onPress={() => {
        void navigate({
          search: openPanel(member ? { profile: member.userId } : { group: group?.id }),
          to: ".",
        });
      }}
    >
      {label}
    </Button>
  );
};
