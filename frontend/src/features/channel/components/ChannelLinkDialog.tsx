import { useState } from "react";

import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { Dialog } from "#/components/ui/Dialog/Dialog";
import { TextField } from "#/components/ui/TextField/TextField";
import { toast } from "#/components/ui/ToastRegion/toast";

import { useChannelLinkActions } from "../hooks/useChannelLinks";

import type { ChannelLink } from "#/gen/chat/v1/channel_link_service_pb";

type ChannelLinkDialogProps = {
  channelId: string;
  // null なら追加、あればそのリンクを編集する
  link: ChannelLink | null;
  onClose: () => void;
};

const hostOf = (url: string) => {
  try {
    const { protocol, host } = new URL(url);
    return protocol === "http:" || protocol === "https:" ? host : null;
  } catch {
    return null;
  }
};

export const ChannelLinkDialog = ({ channelId, link, onClose }: ChannelLinkDialogProps) => {
  const { t } = useTranslation();
  const [url, setUrl] = useState(link?.url ?? "https://");
  const [title, setTitle] = useState(link?.title ?? "");
  const [isSubmitted, setIsSubmitted] = useState(false);
  const { create, update, remove } = useChannelLinkActions(channelId);
  const host = hostOf(url.trim());
  const failed = [create, update, remove].find((mutation) => mutation.isError);

  const save = () => {
    setIsSubmitted(true);
    if (host === null) {
      return;
    }
    const input = { title: title.trim() || host, url: url.trim() };
    const onSuccess = () => {
      toast(t(link ? "channel.links.updated" : "channel.links.added"));
      onClose();
    };
    if (link) {
      update.mutate({ ...input, linkId: link.id }, { onSuccess });
    } else {
      create.mutate({ ...input, channelId }, { onSuccess });
    }
  };

  return (
    <Dialog
      isOpen
      onOpenChange={(isOpen) => {
        if (!isOpen) {
          onClose();
        }
      }}
      title={t(link ? "channel.links.editTitle" : "channel.links.addTitle")}
      footer={
        <>
          {link && (
            <Button
              variant="ghost"
              className="mr-auto text-danger"
              isPending={remove.isPending}
              onPress={() => {
                remove.mutate(
                  { linkId: link.id },
                  {
                    onSuccess: () => {
                      toast(t("channel.links.deleted"));
                      onClose();
                    },
                  },
                );
              }}
            >
              {t("common.delete")}
            </Button>
          )}
          <Button variant="secondary" onPress={onClose}>
            {t("common.cancel")}
          </Button>
          <Button isPending={create.isPending || update.isPending} onPress={save}>
            {t(link ? "common.save" : "channel.links.add")}
          </Button>
        </>
      }
    >
      <p className="m-0 text-caption text-muted">{t("channel.links.hint")}</p>
      <TextField
        label={t("channel.links.url")}
        type="url"
        value={url}
        onChange={setUrl}
        errorMessage={isSubmitted && host === null ? t("channel.links.invalidUrl") : undefined}
        maxLength={2048}
      />
      <TextField
        label={t("channel.links.title")}
        placeholder={host ?? t("channel.links.titlePlaceholder")}
        description={t("channel.links.titleHint")}
        value={title}
        onChange={setTitle}
        maxLength={100}
      />
      {failed && <p className="m-0 text-caption text-danger">{failed.error.message}</p>}
    </Dialog>
  );
};
