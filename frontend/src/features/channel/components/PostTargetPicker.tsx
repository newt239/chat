import { IconCheck, IconChevronDown, IconHash, IconLock } from "@tabler/icons-react";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Menu } from "#/components/ui/Menu";
import { MenuItem } from "#/components/ui/MenuItem";
import { focusRing } from "#/components/ui/styles";

import { lastSegment } from "../utils/channelPath";
import { relativePath } from "../utils/channelTree";

import type { Channel } from "#/gen/chat/v1/channel_service_pb";

type PostTargetPickerProps = {
  parent: Channel;
  descendants: Channel[];
  value: string;
  onChange: (channelId: string) => void;
};

// 集約表示中の投稿先の切り替え。初期値は親チャンネル
export const PostTargetPicker = ({
  parent,
  descendants,
  value,
  onChange,
}: PostTargetPickerProps) => {
  const { t } = useTranslation();
  const labelOf = (channel: Channel) =>
    channel.id === parent.id
      ? t("channel.aggregate.thisChannel", { name: lastSegment(parent.name) })
      : relativePath(parent.name, channel.name);
  const selected = [parent, ...descendants].find((channel) => channel.id === value) ?? parent;

  return (
    <Menu
      placement="top start"
      trigger={
        <Button
          aria-label={t("channel.aggregate.target", { name: selected.name })}
          className={`mb-1.5 inline-flex h-[26px] cursor-pointer items-center gap-1 rounded-[6px] border border-border px-2 text-xs whitespace-nowrap text-muted data-hovered:border-border-strong [&_svg]:size-3 ${focusRing}`}
        >
          {t("channel.aggregate.targetLabel")}
          <b className="font-semibold text-text">
            #{" "}
            {selected.id === parent.id
              ? lastSegment(parent.name)
              : relativePath(parent.name, selected.name)}
          </b>
          <IconChevronDown aria-hidden />
        </Button>
      }
    >
      {[parent, ...descendants].map((channel) => (
        <MenuItem
          key={channel.id}
          icon={
            channel.id === value ? <IconCheck /> : channel.isPrivate ? <IconLock /> : <IconHash />
          }
          onAction={() => {
            onChange(channel.id);
          }}
        >
          {labelOf(channel)}
        </MenuItem>
      ))}
    </Menu>
  );
};
