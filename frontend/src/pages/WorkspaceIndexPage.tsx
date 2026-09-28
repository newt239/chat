import { IconHash } from "@tabler/icons-react";
import { Navigate, useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { EmptyState } from "#/components/ui/EmptyState";
import { useChannels } from "#/features/channel/hooks/useChannel";

// 参加しているチャンネルがあれば先頭を開く
export const WorkspaceIndexPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const { data: channels } = useChannels(workspaceId);
  const first = channels
    ?.filter((channel) => channel.isMember)
    .toSorted((a, b) => a.name.localeCompare(b.name))[0];

  if (first) {
    return (
      <Navigate
        to="/app/$workspaceId/$channelId"
        params={{ channelId: first.id, workspaceId }}
        replace
      />
    );
  }

  return (
    <EmptyState
      icon={<IconHash />}
      title={t("shell.workspaceIndex.title")}
      description={t("shell.workspaceIndex.description")}
    />
  );
};
