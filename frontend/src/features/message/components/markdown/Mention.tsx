import type { ReactNode } from "react";

import { Badge } from "@mantine/core";
import { useParams } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { useMembers } from "#/features/member/hooks/useMembers";
import { setRightSidePanelViewAtom } from "#/providers/store/ui";

type MentionProps = {
  "data-mention": string;
  children?: ReactNode;
};

export const Mention = ({ "data-mention": username }: MentionProps) => {
  const { workspaceId } = useParams({ strict: false });
  const { data: members } = useMembers(workspaceId ?? null);
  const setRightSidePanelView = useSetAtom(setRightSidePanelViewAtom);

  // メンションは表示名の前方一致で解決される
  const member = members?.find((item) => item.displayName.startsWith(username));

  return (
    <Badge
      variant="light"
      color="blue"
      size="sm"
      className={member === undefined ? "" : "cursor-pointer hover:bg-blue-100"}
      component="span"
      onClick={() => {
        if (member !== undefined) {
          setRightSidePanelView({ type: "user-profile", userId: member.userId });
        }
      }}
    >
      @{username}
    </Badge>
  );
};
