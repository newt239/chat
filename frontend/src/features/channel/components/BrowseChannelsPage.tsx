import { useState } from "react";

import { useQuery } from "@connectrpc/connect-query";
import { IconHash, IconSearch } from "@tabler/icons-react";
import { getRouteApi } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { PageHeader } from "#/components/block/PageHeader/PageHeader";
import { Badge } from "#/components/ui/Badge/Badge";
import { Button } from "#/components/ui/Button/Button";
import { EmptyState } from "#/components/ui/EmptyState/EmptyState";
import { Link } from "#/components/ui/Link/Link";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { TextField } from "#/components/ui/TextField/TextField";
import { toast } from "#/components/ui/ToastRegion/toast";
import { ChannelService } from "#/gen/chat/v1/channel_service_pb";

import { useChannelMemberActions } from "../hooks/useChannelMemberActions";
import { ChannelName } from "./ChannelName";

const routeApi = getRouteApi("/app/$workspaceId/browse-channels");

// 参加していないチャンネルも含めて一覧し、開くとプレビュー、ボタンで参加する
export const BrowseChannelsPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = routeApi.useParams();
  const [query, setQuery] = useState("");
  const { data: channels, isLoading } = useQuery(
    ChannelService.method.listBrowsableChannels,
    { workspaceId },
    { select: (res) => res.channels },
  );
  const { join } = useChannelMemberActions(workspaceId);
  const keyword = query.trim().toLowerCase();
  const matched = (channels ?? []).filter(
    ({ channel }) =>
      channel !== undefined &&
      (channel.name.includes(keyword) ||
        (channel.description ?? "").toLowerCase().includes(keyword)),
  );

  return (
    <>
      <PageHeader icon={<IconHash />} title={t("channel.browse.title")} />
      <div className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto px-[18px] py-3 max-md:px-3">
        <TextField
          label={t("channel.browse.search")}
          type="search"
          value={query}
          onChange={setQuery}
          placeholder={t("channel.browse.search")}
        />
        {isLoading && <Skeleton className="h-12 w-full" />}
        {!isLoading && matched.length === 0 && (
          <EmptyState icon={<IconSearch />} title={t("channel.browse.empty")} description={query} />
        )}
        <ul className="m-0 flex list-none flex-col p-0">
          {matched.map(({ channel, memberCount }) =>
            channel === undefined ? null : (
              <li
                key={channel.id}
                className="flex items-center gap-3 border-b border-border py-2.5 last:border-b-0"
              >
                <div className="flex min-w-0 flex-1 flex-col gap-0.5">
                  <Link
                    to="/app/$workspaceId/$channelId"
                    params={{ channelId: channel.id, workspaceId }}
                    className="flex min-w-0 items-center gap-1 text-[14px] font-semibold text-text no-underline data-hovered:underline [&_svg]:size-4 [&_svg]:shrink-0 [&_svg]:text-muted"
                  >
                    <ChannelName name={channel.name} isPrivate={channel.isPrivate} />
                  </Link>
                  <span className="truncate text-caption text-muted">
                    {t("channel.browse.memberCount", { count: memberCount })}
                    {" · "}
                    {channel.description || t("channel.browse.noDescription")}
                  </span>
                </div>
                {channel.isMember ? (
                  <Badge tone="accent">{t("channel.browse.joined")}</Badge>
                ) : (
                  <Button
                    size="sm"
                    variant="secondary"
                    isPending={join.isPending && join.variables.channelId === channel.id}
                    onPress={() => {
                      join.mutate(
                        { channelId: channel.id },
                        {
                          onSuccess: () => {
                            toast(t("channel.browse.joinedToast", { name: channel.name }), {
                              tone: "success",
                            });
                          },
                        },
                      );
                    }}
                  >
                    {t("channel.browse.join")}
                  </Button>
                )}
              </li>
            ),
          )}
        </ul>
      </div>
    </>
  );
};
