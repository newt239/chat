import { IconHash, IconSlash, IconSpeakerphone, IconUsersGroup } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar/Avatar";
import { cn } from "#/components/ui/styles/styles";
import { commandNames } from "#/features/command/utils/commands";

import type { SuggestionItem } from "../utils/suggestion";

type SuggestionListProps = {
  id: string;
  items: SuggestionItem[];
  activeIndex: number;
  onSelect: (item: SuggestionItem) => void;
};

// 入力欄の上に出す @ / # / の候補。フォーカスは入力欄に残し、aria-activedescendant で選択中を伝える
export const SuggestionList = ({ id, items, activeIndex, onSelect }: SuggestionListProps) => {
  const { t } = useTranslation();

  return (
    <div
      id={id}
      role="listbox"
      aria-label={t("message.suggestion.label")}
      className="absolute bottom-full left-0 z-10 mb-1 flex max-h-72 w-full max-w-md flex-col overflow-y-auto rounded-lg border border-border bg-surface p-1 shadow-lg"
    >
      {items.map((item, index) => (
        <div
          key={`${item.kind}-${item.id}`}
          id={`${id}-${index}`}
          role="option"
          aria-selected={index === activeIndex}
          // 入力欄からフォーカスを奪わないよう、押した時点で選ぶ
          onPointerDown={(event) => {
            event.preventDefault();
            onSelect(item);
          }}
          className={cn(
            "flex cursor-pointer items-center gap-2 rounded-md px-2 py-1.5 text-[13px] [&_svg]:size-4 [&_svg]:text-muted",
            index === activeIndex && "bg-hover",
          )}
        >
          {item.kind === "user" && <Avatar name={item.label} src={item.avatarUrl} size={20} />}
          {item.kind === "group" && <IconUsersGroup aria-hidden />}
          {item.kind === "channel" && <IconHash aria-hidden />}
          {item.kind === "broadcast" && <IconSpeakerphone aria-hidden />}
          {item.kind === "command" && <IconSlash aria-hidden />}
          <span
            className={cn(
              "min-w-0 flex-1 truncate font-medium",
              item.kind === "command" && "flex-none",
            )}
          >
            {item.label}
          </span>
          {(item.kind === "group" || item.kind === "broadcast") && (
            <span className="shrink-0 text-caption text-subtle">
              {t(
                item.kind === "group"
                  ? "message.suggestion.groups"
                  : "message.suggestion.broadcast",
              )}
            </span>
          )}
          {item.kind === "command" && (
            <span className="min-w-0 flex-1 truncate text-caption text-subtle">
              {commandNames
                .filter((name) => `/${name}` === item.value)
                .map((name) => t(`command.${name}.usage`))}
            </span>
          )}
        </div>
      ))}
    </div>
  );
};
