import { getRouteApi } from "@tanstack/react-router";
import { Button, ToggleButton } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { focusRing } from "#/components/ui/styles/styles";
import { chipClassName } from "#/features/search/utils/chipClassName";
import {
  emptySearchQuery,
  formatSearchQuery,
  hasSearchConditions,
  searchHasValues,
  searchIsValues,
} from "#/features/search/utils/searchQuery";

import { SearchDateFilter } from "./SearchDateFilter";
import { SearchFilterPicker } from "./SearchFilterPicker";

import type { ResolvedSearchQuery } from "#/features/search/hooks/useWorkspaceSearch";
import type { SearchParams } from "#/features/search/schemas";
import type { SearchQuery } from "#/features/search/utils/searchQuery";

const searchRoute = getRouteApi("/app/$workspaceId/search");

const toggle = <T,>(list: readonly T[], value: T) =>
  list.includes(value) ? list.filter((item) => item !== value) : [...list, value];

// 選んだ候補に当たる名前があれば外し、なければ足す
const toggleName = (names: readonly string[], matched: readonly string[], added: string) =>
  matched.length > 0 ? names.filter((name) => !matched.includes(name)) : [...names, added];

const summarize = (labels: readonly string[]) =>
  labels.length === 0
    ? null
    : labels.length === 1
      ? (labels[0] ?? null)
      : `${labels[0]} +${labels.length - 1}`;

type SearchFilterBarProps = {
  resolved: ResolvedSearchQuery;
  // 解決できなかった修飾子（@名前 / #チャンネル名）
  unresolved: string[];
};

// 入力欄の修飾子と同じ条件をチップで操作する。操作すると q を書き換える
export const SearchFilterBar = ({ resolved, unresolved }: SearchFilterBarProps) => {
  const { t } = useTranslation();
  const search = searchRoute.useSearch();
  const navigate = searchRoute.useNavigate();
  const { query, users, inChannels, members, channels, hasDescendants } = resolved;

  const update = (patch: Partial<SearchParams>) => {
    void navigate({ search: (prev) => ({ ...prev, ...patch, page: 1 }) });
  };
  const setQuery = (patch: Partial<SearchQuery>) => {
    update({ q: formatSearchQuery({ ...query, ...patch }) });
  };

  const hasConditions = hasSearchConditions(query) || !search.replies;

  return (
    <div
      role="group"
      aria-label={t("search.filters.label")}
      className="flex shrink-0 gap-1.5 overflow-x-auto px-4.5 pb-2.5 [scrollbar-width:none] md:flex-wrap"
    >
      <SearchFilterPicker
        label={t("search.filters.from")}
        summary={summarize(users.map(({ member, name }) => member?.displayName ?? name))}
        options={members.map((member) => ({ label: member.displayName, value: member.userId }))}
        selected={users.flatMap(({ member }) => (member ? [member.userId] : []))}
        onToggle={(userId) => {
          setQuery({
            from: toggleName(
              query.from,
              users.flatMap(({ member, name }) => (member?.userId === userId ? [name] : [])),
              members.find((item) => item.userId === userId)?.displayName ?? userId,
            ),
          });
        }}
        isSearchable
        isInvalid={unresolved.some((name) => name.startsWith("@"))}
      />
      <SearchFilterPicker
        label={t("search.filters.in")}
        summary={summarize(inChannels.map(({ name }) => `#${name}`))}
        options={channels.map((channel) => ({ label: `#${channel.name}`, value: channel.id }))}
        selected={inChannels.flatMap(({ channel }) => (channel ? [channel.id] : []))}
        onToggle={(channelId) => {
          setQuery({
            in: toggleName(
              query.in,
              inChannels.flatMap(({ channel, name }) => (channel?.id === channelId ? [name] : [])),
              channels.find((item) => item.id === channelId)?.name ?? channelId,
            ),
          });
        }}
        isSearchable
        isInvalid={unresolved.some((name) => name.startsWith("#"))}
      />
      {hasDescendants && (
        <ToggleButton
          isSelected={search.subs}
          onChange={(subs) => {
            update({ subs });
          }}
          className={chipClassName}
        >
          {t("search.filters.subs")}
        </ToggleButton>
      )}
      <SearchDateFilter
        after={query.after}
        before={query.before}
        isInvalid={query.invalidDates.length > 0}
        onChange={(range) => {
          setQuery({ ...range, invalidDates: [] });
        }}
      />
      <SearchFilterPicker
        label={t("search.filters.has")}
        summary={summarize(query.has.map((has) => t(`search.has.${has}`)))}
        options={searchHasValues.map((value) => ({ label: t(`search.has.${value}`), value }))}
        selected={query.has}
        onToggle={(has) => {
          setQuery({ has: toggle(query.has, has) });
        }}
        isSearchable={false}
        isInvalid={false}
      />
      {searchIsValues.map((is) => (
        <ToggleButton
          key={is}
          isSelected={query.is.includes(is)}
          onChange={() => {
            setQuery({ is: toggle(query.is, is) });
          }}
          className={chipClassName}
        >
          {t(`search.filters.${is}`)}
        </ToggleButton>
      ))}
      <ToggleButton
        isSelected={search.replies}
        onChange={(replies) => {
          update({ replies });
        }}
        className={chipClassName}
      >
        {t("search.filters.replies")}
      </ToggleButton>
      {hasConditions && (
        <Button
          onPress={() => {
            update({
              q: formatSearchQuery({ ...emptySearchQuery, keywords: query.keywords }),
              replies: true,
              subs: true,
            });
          }}
          className={`h-7.5 shrink-0 cursor-pointer rounded-md px-2 font-sans text-xs font-semibold whitespace-nowrap text-accent-text data-hovered:bg-hover ${focusRing}`}
        >
          {t("search.filters.clear")}
        </Button>
      )}
    </div>
  );
};
