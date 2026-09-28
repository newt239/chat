import { useState } from "react";

import { IconArrowDown, IconArrowUp, IconEdit, IconLink } from "@tabler/icons-react";
import { Link } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button";
import { IconButton } from "#/components/ui/IconButton";
import { focusRing } from "#/components/ui/styles";

import { useChannelLinkActions, useChannelLinks } from "../hooks/useChannelLinks";
import { ChannelLinkDialog } from "./ChannelLinkDialog";

import type { ChannelLink } from "#/gen/chat/v1/channel_link_service_pb";

type ChannelLinksSectionProps = {
  channelId: string;
};

// チャンネル情報の関連リンク。編集できる人は追加・編集・削除と上下の並び替えができる
export const ChannelLinksSection = ({ channelId }: ChannelLinksSectionProps) => {
  const { t } = useTranslation();
  const { data } = useChannelLinks(channelId);
  const { move } = useChannelLinkActions(channelId);
  // undefined は閉じている、null は追加
  const [editing, setEditing] = useState<ChannelLink | null | undefined>(undefined);
  const links = data?.links ?? [];
  const linkIds = links.map((link) => link.id);
  const canEdit = data?.canEdit ?? false;

  return (
    <section className="flex flex-col gap-1.5 border-b border-border px-4 py-3">
      <h4 className="m-0 flex items-center justify-between text-xs font-semibold text-muted">
        {t("shell.channel.links")}
        {canEdit && (
          <Button
            size="sm"
            variant="ghost"
            onPress={() => {
              setEditing(null);
            }}
          >
            {t("channel.links.add")}
          </Button>
        )}
      </h4>
      {links.length === 0 && (
        <p className="m-0 text-[12.5px] text-muted">{t("channel.links.empty")}</p>
      )}
      <ul className="m-0 -mx-2 flex list-none flex-col p-0">
        {links.map((link, index) => (
          <li
            key={link.id}
            className="flex items-center gap-2 rounded-md px-2 py-1 [&_svg]:shrink-0"
          >
            <IconLink aria-hidden className="size-4 text-muted" />
            <span className="flex min-w-0 flex-1 flex-col leading-[1.35]">
              <Link
                href={link.url}
                target="_blank"
                rel="noopener noreferrer"
                className={`truncate rounded-sm text-[13.5px] text-text no-underline data-hovered:underline ${focusRing}`}
              >
                {link.title}
              </Link>
              <small className="truncate text-[11.5px] text-subtle">{link.url}</small>
            </span>
            {canEdit && (
              <span className="flex shrink-0 [&_button]:size-7 [&_svg]:size-4!">
                <IconButton
                  label={t("channel.links.moveUp", { title: link.title })}
                  isDisabled={index === 0}
                  onPress={() => {
                    move(linkIds, index, index - 1);
                  }}
                >
                  <IconArrowUp />
                </IconButton>
                <IconButton
                  label={t("channel.links.moveDown", { title: link.title })}
                  isDisabled={index === links.length - 1}
                  onPress={() => {
                    move(linkIds, index, index + 1);
                  }}
                >
                  <IconArrowDown />
                </IconButton>
                <IconButton
                  label={t("channel.links.editOf", { title: link.title })}
                  onPress={() => {
                    setEditing(link);
                  }}
                >
                  <IconEdit />
                </IconButton>
              </span>
            )}
          </li>
        ))}
      </ul>
      {editing !== undefined && (
        <ChannelLinkDialog
          channelId={channelId}
          link={editing}
          isOpen
          onOpenChange={(isOpen) => {
            if (!isOpen) {
              setEditing(undefined);
            }
          }}
        />
      )}
    </section>
  );
};
