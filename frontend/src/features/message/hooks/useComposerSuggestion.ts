import { useEffect, useId, useState } from "react";

import { skipToken, useQuery } from "@connectrpc/connect-query";
import { useParams } from "@tanstack/react-router";

import { useMembers } from "#/features/member/hooks/useMembers";
import { useUserGroups } from "#/features/userGroup/hooks/useUserGroups";
import { ChannelService } from "#/gen/chat/v1/channel_service_pb";

import {
  applySuggestion,
  findSuggestionQuery,
  mentionTokenOf,
  rankByQuery,
} from "../utils/suggestion";

import type { SuggestionItem } from "../utils/suggestion";

type Options = {
  body: string;
  cursor: number;
  onApply: (next: { text: string; cursor: number }) => void;
};

type SuggestionKeyEvent = {
  key: string;
  nativeEvent: { isComposing: boolean };
  preventDefault: () => void;
};

const LIMIT = 8;

/** 入力欄で @ を打つとユーザーとユーザーグループ、# を打つとチャンネルの候補を出す */
export const useComposerSuggestion = ({ body, cursor, onApply }: Options) => {
  const { workspaceId = null } = useParams({ strict: false });
  const listId = useId();
  const query = findSuggestionQuery(body, cursor);
  const [activeIndex, setActiveIndex] = useState(0);
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
    query?.trigger === "#"
      ? channels.map((channel) => ({
          avatarUrl: undefined,
          id: channel.id,
          kind: "channel",
          label: channel.name,
          value: `#${channel.name}`,
        }))
      : [
          ...members.flatMap((member) => {
            const token = mentionTokenOf(member.displayName);
            return token === null || member.suspendedAt !== undefined
              ? []
              : [
                  {
                    avatarUrl: member.avatarUrl,
                    id: member.userId,
                    kind: "user" as const,
                    label: member.nickname ?? member.displayName,
                    value: `@${token}`,
                  },
                ];
          }),
          ...groups.map((group) => ({
            avatarUrl: undefined,
            id: group.id,
            kind: "group" as const,
            label: group.name,
            value: `@${group.name}`,
          })),
        ];
  const items =
    query === null
      ? []
      : rankByQuery(candidates, query.query, (item) => item.label).slice(0, LIMIT);
  const isOpen = query !== null && items.length > 0 && dismissedStart !== query.start;
  const active = Math.min(activeIndex, items.length - 1);

  useEffect(() => {
    setActiveIndex(0);
  }, [query?.start, query?.query]);

  const select = (item: SuggestionItem) => {
    if (query !== null) {
      onApply(applySuggestion({ cursor, query, text: body, value: item.value }));
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
