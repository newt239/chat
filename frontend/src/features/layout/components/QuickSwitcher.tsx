import { useState } from "react";

import { IconSearch, IconUser, IconUsers } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { Autocomplete, Menu, MenuItem, useFilter } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Dialog } from "#/components/ui/Dialog/Dialog";
import { SearchField } from "#/components/ui/SearchField/SearchField";
import { cn, overlayStyles } from "#/components/ui/styles/styles";
import { ChannelName } from "#/features/channel/components/ChannelName";
import { useChannels } from "#/features/channel/hooks/useChannel";
import { useDMs } from "#/features/channel/hooks/useDM";
import { dmName } from "#/features/channel/utils/dmName";
import { useDisplayName } from "#/features/member/hooks/useDisplayName";

type QuickSwitcherProps = {
  workspaceId: string;
  onClose: () => void;
};

const itemClassName = cn(overlayStyles.listItem, "[&_svg]:size-4 [&_svg]:shrink-0");

// ⌘K で開く移動先の絞り込み。チャンネルと DM を名前で探し、見つからなければメッセージ検索へ渡す
export const QuickSwitcher = ({ workspaceId, onClose }: QuickSwitcherProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { contains } = useFilter({ sensitivity: "base" });
  const [query, setQuery] = useState("");
  const { data: channels = [] } = useChannels(workspaceId);
  const { data: dms = [] } = useDMs(workspaceId);
  const displayName = useDisplayName();

  const open = (channelId: string) => {
    void navigate({ params: { channelId, workspaceId }, to: "/app/$workspaceId/$channelId" });
  };

  return (
    <Dialog isOpen onOpenChange={onClose} title={t("shell.switcher.title")}>
      <Autocomplete inputValue={query} onInputChange={setQuery} filter={contains}>
        <SearchField label={t("shell.switcher.placeholder")} autoFocus />
        <Menu
          aria-label={t("shell.switcher.title")}
          renderEmptyState={() => (
            <p className="m-0 px-2.5 py-1.5 text-body-sm text-muted">{t("shell.switcher.empty")}</p>
          )}
          className="-mx-1 max-h-90 overflow-y-auto pb-2 outline-none"
        >
          {channels.map((channel) => (
            <MenuItem
              key={channel.id}
              id={channel.id}
              textValue={channel.name}
              onAction={() => {
                open(channel.id);
              }}
              className={itemClassName}
            >
              <ChannelName name={channel.name} isPrivate={channel.isPrivate} />
            </MenuItem>
          ))}
          {dms.map((dm) => {
            const name = dmName(dm, displayName);
            const Icon = dm.members.length > 1 ? IconUsers : IconUser;
            return (
              <MenuItem
                key={dm.id}
                id={dm.id}
                textValue={name}
                onAction={() => {
                  open(dm.id);
                }}
                className={itemClassName}
              >
                <Icon aria-hidden />
                <span className="min-w-0 flex-1 truncate">{name}</span>
              </MenuItem>
            );
          })}
          {query !== "" && (
            // textValue を入力そのものにして、絞り込みで常に残す
            <MenuItem
              id="search"
              textValue={query}
              onAction={() => {
                void navigate({
                  params: { workspaceId },
                  search: { q: query },
                  to: "/app/$workspaceId/search",
                });
              }}
              className={itemClassName}
            >
              <IconSearch aria-hidden />
              <span className="min-w-0 flex-1 truncate">
                {t("shell.switcher.searchMessages", { query })}
              </span>
            </MenuItem>
          )}
        </Menu>
      </Autocomplete>
    </Dialog>
  );
};
