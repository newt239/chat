import { Text, UnstyledButton } from "@mantine/core";
import { IconUser, IconUsers } from "@tabler/icons-react";
import { Link } from "react-router";

import { paths } from "#/lib/paths";
import { useOptionalRouteParams } from "#/lib/routeParams";

import { useDMs } from "../hooks/useDM";

import type { components } from "#/lib/api/schema";

type DM = components["schemas"]["DMOutput"];

const getDMDisplayName = (dm: DM) => {
  if (dm.type === "dm") {
    const [otherMember] = dm.members;
    return otherMember?.displayName || "不明なユーザー";
  }
  return dm.name || `グループDM (${dm.members.length}人)`;
};

type DMListProps = {
  workspaceId: string;
};

export const DMList = ({ workspaceId }: DMListProps) => {
  const { channelId } = useOptionalRouteParams();
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
          <Link key={dm.id} to={paths.channel(workspaceId, dm.id)} className="block no-underline">
            <UnstyledButton
              className={`w-full px-3 py-1.5 rounded-md flex items-center space-x-2 transition-colors ${
                isActive ? "bg-blue-50 text-blue-600" : "text-gray-700 hover:bg-gray-100"
              }`}
            >
              {dm.type === "dm" ? <IconUser size={16} /> : <IconUsers size={16} />}
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
