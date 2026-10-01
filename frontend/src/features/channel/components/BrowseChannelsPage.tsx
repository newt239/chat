import { useQuery } from "@connectrpc/connect-query";
import { IconChevronLeft, IconChevronRight, IconHash, IconSearch } from "@tabler/icons-react";
import { keepPreviousData } from "@tanstack/react-query";
import { getRouteApi } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { PageHeader } from "#/components/block/PageHeader/PageHeader";
import { Badge } from "#/components/ui/Badge/Badge";
import { Button } from "#/components/ui/Button/Button";
import { EmptyState } from "#/components/ui/EmptyState/EmptyState";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { Link } from "#/components/ui/Link/Link";
import { SegmentedControl } from "#/components/ui/SegmentedControl/SegmentedControl";
import { Select } from "#/components/ui/Select/Select";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { TextField } from "#/components/ui/TextField/TextField";
import { toast } from "#/components/ui/ToastRegion/toast";
import { ChannelService } from "#/gen/chat/v1/channel_service_pb";

import { useChannelMemberActions } from "../hooks/useChannelMemberActions";
import {
  browseMembershipMessages,
  browseMembershipValues,
  browseSortMessages,
  browseSortValues,
} from "../schemas";
import { ChannelName } from "./ChannelName";

const routeApi = getRouteApi("/app/$workspaceId/browse-channels");
const PER_PAGE = 20;

// 参加していないチャンネルも含めて一覧し、開くとプレビュー、ボタンで参加する
export const BrowseChannelsPage = () => {
  const { t } = useTranslation();
  const { workspaceId } = routeApi.useParams();
  const { membership, page, q: query, sort } = routeApi.useSearch();
  const navigate = routeApi.useNavigate();
  const { data, isLoading } = useQuery(
    ChannelService.method.searchBrowsableChannels,
    {
      membership: browseMembershipMessages[membership],
      page,
      perPage: PER_PAGE,
      query: query.trim(),
      sort: browseSortMessages[sort],
      workspaceId,
    },
    { placeholderData: keepPreviousData },
  );
  const { join } = useChannelMemberActions(workspaceId);
  const matched = data?.channels ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / PER_PAGE));
  // 条件を変えたら 1 ページ目に戻す
  const changeFilter = (
    next: Partial<{ membership: typeof membership; q: string; sort: typeof sort }>,
  ) => {
    void navigate({ replace: true, search: (prev) => ({ ...prev, ...next, page: 1 }) });
  };

  return (
    <>
      <PageHeader icon={<IconHash />} title={t("channel.browse.title")} />
      <div className="flex min-h-0 flex-1 flex-col gap-3 overflow-y-auto px-[18px] py-3 max-md:px-3">
        <TextField
          label={t("channel.browse.search")}
          type="search"
          value={query}
          onChange={(q) => {
            changeFilter({ q });
          }}
          placeholder={t("channel.browse.search")}
        />
        <div className="flex flex-wrap items-center justify-between gap-2">
          <SegmentedControl
            label={t("channel.browse.membership.label")}
            options={browseMembershipValues.map((value) => ({
              label: t(`channel.browse.membership.${value}`),
              value,
            }))}
            value={membership}
            onChange={(next) => {
              changeFilter({ membership: next });
            }}
          />
          <Select
            label={t("channel.browse.sort.label")}
            options={browseSortValues.map((value) => ({
              label: t(`channel.browse.sort.${value}`),
              value,
            }))}
            value={sort}
            onChange={(next) => {
              changeFilter({ sort: next });
            }}
            className="flex-row items-center gap-2 [&_button]:h-7 [&_button]:w-40 [&_button]:text-[12.5px] [&_label]:text-xs [&_label]:font-normal [&_label]:text-muted"
          />
        </div>
        {data !== undefined && (
          <p className="m-0 text-xs text-muted">{t("channel.browse.count", { count: total })}</p>
        )}
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
        {totalPages > 1 && (
          <nav className="flex items-center justify-center gap-2 pb-3 text-caption text-muted">
            <IconButton
              label={t("channel.browse.prev")}
              isDisabled={page <= 1}
              onPress={() => {
                void navigate({ search: (prev) => ({ ...prev, page: page - 1 }) });
              }}
            >
              <IconChevronLeft />
            </IconButton>
            <span className="tabular-nums">
              {t("channel.browse.page", { page, total: totalPages })}
            </span>
            <IconButton
              label={t("channel.browse.next")}
              isDisabled={page >= totalPages}
              onPress={() => {
                void navigate({ search: (prev) => ({ ...prev, page: page + 1 }) });
              }}
            >
              <IconChevronRight />
            </IconButton>
          </nav>
        )}
      </div>
    </>
  );
};
