import { useId, useState } from "react";

import { useMentionDirectory } from "#/features/mention/hooks/useMentionDirectory";
import { toMentionToken } from "#/features/mention/utils/mentionToken";
import { commandNames } from "#/features/message/utils/commands";

import { applySuggestion, findSuggestionQuery, rankByQuery } from "../utils/suggestion";

import type { SuggestionItem } from "../utils/suggestion";

import type { MentionToken } from "#/features/mention/utils/mentionToken";

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

const mentionItem = (token: MentionToken, label: string, avatarUrl: string | undefined) => ({
  avatarUrl,
  id: token.id,
  kind: token.kind,
  label,
  token: toMentionToken(token),
  value: `${token.kind === "channel" ? "#" : "@"}${label}`,
});

const broadcasts = (["channel", "here"] as const).map((id) =>
  mentionItem({ id, kind: "broadcast" }, id, undefined),
);

/** 入力欄で @ を打つとユーザーとユーザーグループ、# を打つとチャンネル、先頭で / を打つとコマンドの候補を出す */
export const useComposerSuggestion = ({ body, cursor, allowsCommands, onApply }: Options) => {
  const { members = [], groups = [], browsable = [] } = useMentionDirectory();
  const listId = useId();
  const query = findSuggestionQuery(body, cursor, allowsCommands);
  // 選択中の位置は検索語ごとに持ち、検索語が変わったら先頭に戻す
  const queryKey = query === null ? null : `${query.start}:${query.query}`;
  const [activeState, setActiveState] = useState({ index: 0, queryKey });
  // Esc で閉じた候補は、別の @ / # を打つまで出さない
  const [dismissedStart, setDismissedStart] = useState<number | null>(null);

  const candidates: SuggestionItem[] =
    query?.trigger === "/"
      ? commands
      : query?.trigger === "#"
        ? browsable.map((channel) =>
            mentionItem({ id: channel.id, kind: "channel" }, channel.name, undefined),
          )
        : [
            ...members
              .filter((member) => member.suspendedAt === undefined)
              .map((member) =>
                mentionItem(
                  { id: member.userId, kind: "user" },
                  member.nickname ?? member.displayName,
                  member.avatarUrl,
                ),
              ),
            ...groups.map((group) =>
              mentionItem({ id: group.id, kind: "group" }, group.name, undefined),
            ),
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
