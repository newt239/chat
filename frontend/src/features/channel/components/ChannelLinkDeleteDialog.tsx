import { useTranslation } from "react-i18next";

import { AlertDialog } from "#/components/ui/AlertDialog/AlertDialog";
import { toast } from "#/components/ui/ToastRegion/toast";

import { useChannelLinkActions } from "../hooks/useChannelLinks";

import type { ChannelLink } from "#/gen/chat/v1/channel_link_service_pb";

type ChannelLinkDeleteDialogProps = {
  channelId: string;
  // null なら閉じている
  link: ChannelLink | null;
  onClose: () => void;
  onDeleted: () => void;
};

export const ChannelLinkDeleteDialog = ({
  channelId,
  link,
  onClose,
  onDeleted,
}: ChannelLinkDeleteDialogProps) => {
  const { t } = useTranslation();
  const { remove } = useChannelLinkActions(channelId);
  return (
    <AlertDialog
      isOpen={link !== null}
      onOpenChange={onClose}
      title={t("channel.links.deleteTitle", { title: link?.title ?? "" })}
      confirmLabel={t("common.delete")}
      tone="danger"
      isPending={remove.isPending}
      onConfirm={() => {
        if (link === null) {
          return;
        }
        remove.mutate(
          { linkId: link.id },
          {
            onSuccess: () => {
              toast(t("channel.links.deleted"));
              onDeleted();
            },
          },
        );
      }}
    >
      <p className="m-0">{t("channel.links.deleteBody")}</p>
    </AlertDialog>
  );
};
