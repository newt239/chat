import { IconCheck, IconFolder, IconFolderPlus } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { MenuItem } from "#/components/ui/MenuItem/MenuItem";
import { MenuSeparator } from "#/components/ui/MenuSeparator/MenuSeparator";
import { Submenu } from "#/components/ui/Submenu/Submenu";
import { toast } from "#/components/ui/ToastRegion/toast";
import { openDialog } from "#/lib/overlaySearch";

import { useChannelCategories, useChannelCategoryActions } from "../hooks/useChannelCategories";
import { categoryOfChannel } from "../utils/channelTree";

import type { Channel } from "#/gen/chat/v1/channel_service_pb";

type MoveToCategoryMenuProps = {
  workspaceId: string;
  channel: Channel;
  channels: readonly Channel[];
};

// チャンネルの操作メニューに入れる「カテゴリに移動」
export const MoveToCategoryMenu = ({ workspaceId, channel, channels }: MoveToCategoryMenuProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data: categories = [] } = useChannelCategories(workspaceId);
  const { setChannel } = useChannelCategoryActions(workspaceId);
  const current = categoryOfChannel(channel, channels, categories);
  // 割り当てがなければ外しても変わらない。祖先から継承しているときもそのまま
  const isAssigned = categories.some((category) => category.channelIds.includes(channel.id));
  const moveTo = (categoryId: string | undefined, name: string) => {
    setChannel.mutate(
      { categoryId, channelId: channel.id },
      {
        onSuccess: () => {
          toast(t("channel.category.moved", { category: name, channel: channel.name }));
        },
      },
    );
  };
  const checkIcon = (categoryId: string | null) => (
    <IconCheck className={categoryId === current ? "" : "invisible"} />
  );

  return (
    <Submenu label={t("channel.category.moveTo")} icon={<IconFolder />}>
      <MenuItem
        icon={checkIcon(null)}
        isDisabled={!isAssigned}
        onAction={() => {
          moveTo(undefined, t("shell.sidebar.channels"));
        }}
      >
        {t("channel.category.defaultCategory")}
      </MenuItem>
      {categories.map((category) => (
        <MenuItem
          key={category.id}
          icon={checkIcon(category.id)}
          onAction={() => {
            moveTo(category.id, category.name);
          }}
        >
          {category.name}
        </MenuItem>
      ))}
      <MenuSeparator />
      <MenuItem
        icon={<IconFolderPlus />}
        onAction={() => {
          void navigate({
            search: openDialog({ assign: channel.id, dialog: "create-category" }),
            to: ".",
          });
        }}
      >
        {t("channel.category.newCategory")}
      </MenuItem>
    </Submenu>
  );
};
