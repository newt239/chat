import { Text, UnstyledButton } from "@mantine/core";
import { IconUser, IconUsers } from "@tabler/icons-react";
import { Link, useParams } from "@tanstack/react-router";

import { DirectMessageType } from "#/gen/chat/v1/direct_message_service_pb";

import { useDMs } from "../hooks/useDM";

import type { DirectMessage } from "#/gen/chat/v1/direct_message_service_pb";

const getDMDisplayName = (dm: DirectMessage) => {
  if (dm.type === DirectMessageType.DM) {
    const [otherMember] = dm.members;
    return otherMember?.displayName || "不明なユーザー";
  }
  return dm.name || `グループDM (${dm.members.length}人)`;
};

type DMListProps = {
  workspaceId: string;
};

export const DMList = ({ workspaceId }: DMListProps) => {
  const { channelId } = useParams({ strict: false });
  const { data: dms, isLoading } = useDMs(workspaceId);

  if (isLoading) {
    return (
      <div className="px-3 py-2">
        <Text size="sm" c="dimmed">
          読み込み中...
        </Text>
      </div>
    );
  }

  if (!dms || dms.length === 0) {
    return (
      <div className="px-3 py-2">
        <Text size="sm" c="dimmed">
          DMがありません
        </Text>
      </div>
    );
  }

  return (
    <div className="space-y-0.5">
      {dms.map((dm) => {
        const isActive = channelId === dm.id;
        const displayName = getDMDisplayName(dm);

        return (
          <Link
            key={dm.id}
            to="/app/$workspaceId/$channelId"
            params={{ channelId: dm.id, workspaceId }}
            className="block no-underline"
          >
            <UnstyledButton
              className={`w-full px-3 py-1.5 rounded-md flex items-center space-x-2 transition-colors ${
                isActive ? "bg-blue-50 text-blue-600" : "text-gray-700 hover:bg-gray-100"
              }`}
            >
              {dm.type === DirectMessageType.DM ? <IconUser size={16} /> : <IconUsers size={16} />}
              <Text size="sm" truncate className="flex-1">
                {displayName}
              </Text>
            </UnstyledButton>
          </Link>
        );
      })}
    </div>
  );
};
