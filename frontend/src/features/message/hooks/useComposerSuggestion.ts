import { useId, useState } from "react";

import { skipToken, useQuery } from "@connectrpc/connect-query";
import { useParams } from "@tanstack/react-router";

import { commandNames } from "#/features/command/utils/commands";
import { useMembers } from "#/features/member/hooks/useMembers";
import { toMentionToken } from "#/features/mention/utils/mentionToken";
import { useUserGroups } from "#/features/userGroup/hooks/useUserGroups";
import { ChannelService } from "#/gen/chat/v1/channel_service_pb";

import { applySuggestion, findSuggestionQuery, rankByQuery } from "../utils/suggestion";

import type { SuggestionItem } from "../utils/suggestion";

type Options = {
  body: string;
  cursor: number;
  // 入力欄の先頭の「/」でコマンドの候補を出すか。編集では出さない
  allowsCommands: boolean;
  // item は選んだ候補。送信時に名前を ID 記法へ戻すために使う
  onApply: (next: { text: string; cursor: number }, item: SuggestionItem) => void;
};

type SuggestionKeyEvent = {
  key: string;
  nativeEvent: { isComposing: boolean };
  preventDefault: () => void;
};

const LIMIT = 8;

const commands = commandNames.map((name) => ({
  avatarUrl: undefined,
  id: name,
  kind: "command" as const,
  label: `/${name}`,
  token: `/${name}`,
  value: `/${name}`,
}));

const broadcasts = (["channel", "here"] as const).map((id) => ({
  avatarUrl: undefined,
  id,
  kind: "broadcast" as const,
  label: id,
  token: toMentionToken({ id, kind: "broadcast" }),
  value: `@${id}`,
}));

/** 入力欄で @ を打つとユーザーとユーザーグループ、# を打つとチャンネル、先頭で / を打つとコマンドの候補を出す */
export const useComposerSuggestion = ({ body, cursor, allowsCommands, onApply }: Options) => {
  const { workspaceId = null } = useParams({ strict: false });
  const listId = useId();
  const query = findSuggestionQuery(body, cursor, allowsCommands);
  // 選択中の位置は検索語ごとに持ち、検索語が変わったら先頭に戻す
  const queryKey = query === null ? null : `${query.start}:${query.query}`;
  const [activeState, setActiveState] = useState({ index: 0, queryKey });
  // Esc で閉じた候補は、別の @ / # を打つまで出さない
  const [dismissedStart, setDismissedStart] = useState<number | null>(null);
  const { data: members = [] } = useMembers(query?.trigger === "@" ? workspaceId : null);
  const { data: groups = [] } = useUserGroups(query?.trigger === "@" ? workspaceId : null);
  const { data: channels = [] } = useQuery(
    ChannelService.method.listBrowsableChannels,
    query?.trigger === "#" && workspaceId !== null ? { workspaceId } : skipToken,
    { select: (res) => res.channels.flatMap(({ channel }) => channel ?? []) },
  );

  const candidates: SuggestionItem[] =
    query?.trigger === "/"
      ? commands
      : query?.trigger === "#"
        ? channels.map((channel) => ({
            avatarUrl: undefined,
            id: channel.id,
            kind: "channel",
            label: channel.name,
            token: toMentionToken({ id: channel.id, kind: "channel" }),
            value: `#${channel.name}`,
          }))
        : [
            ...members
              .filter((member) => member.suspendedAt === undefined)
              .map((member) => {
                const label = member.nickname ?? member.displayName;
                return {
                  avatarUrl: member.avatarUrl,
                  id: member.userId,
                  kind: "user" as const,
                  label,
                  token: toMentionToken({ id: member.userId, kind: "user" }),
                  value: `@${label}`,
                };
              }),
            ...groups.map((group) => ({
              avatarUrl: undefined,
              id: group.id,
              kind: "group" as const,
              label: group.name,
              token: toMentionToken({ id: group.id, kind: "group" }),
              value: `@${group.name}`,
            })),
            ...broadcasts,
          ];
  const items =
    query === null
      ? []
      : rankByQuery(candidates, query.query, (item) => item.label).slice(0, LIMIT);
  const isOpen = query !== null && items.length > 0 && dismissedStart !== query.start;
  const active = Math.min(
    activeState.queryKey === queryKey ? activeState.index : 0,
    items.length - 1,
  );
  const setActiveIndex = (index: number) => {
    setActiveState({ index, queryKey });
  };

  const select = (item: SuggestionItem) => {
    if (query !== null) {
      onApply(applySuggestion({ cursor, query, text: body, value: item.value }), item);
    }
  };

  // 候補を開いている間は矢印・Enter・Tab・Esc を候補の操作に使う。処理したら true を返す
  const handleKeyDown = (event: SuggestionKeyEvent) => {
    if (!isOpen || event.nativeEvent.isComposing) {
      return false;
    }
    const item = items[active];
    switch (event.key) {
      case "ArrowDown": {
        setActiveIndex((active + 1) % items.length);
        break;
      }
      case "ArrowUp": {
        setActiveIndex((active - 1 + items.length) % items.length);
        break;
      }
      case "Enter":
      case "Tab": {
        if (item !== undefined) {
          select(item);
        }
        break;
      }
      case "Escape": {
        setDismissedStart(query.start);
        break;
      }
      default: {
        return false;
      }
    }
    event.preventDefault();
    return true;
  };

  return {
    handleKeyDown,
    // 入力欄に付ける。スクリーンリーダーに選択中の候補を伝える
    inputProps: isOpen
      ? {
          "aria-activedescendant": `${listId}-${active}`,
          "aria-autocomplete": "list" as const,
          "aria-controls": listId,
          "aria-expanded": true,
        }
      : {},
    isOpen,
    listProps: { activeIndex: active, id: listId, items, onSelect: select },
  };
};
