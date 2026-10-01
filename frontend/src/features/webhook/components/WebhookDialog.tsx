import { useState } from "react";

import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { IconImageField } from "#/components/block/IconImageField/IconImageField";
import { AlertDialog } from "#/components/ui/AlertDialog/AlertDialog";
import { Button } from "#/components/ui/Button/Button";
import { Dialog } from "#/components/ui/Dialog/Dialog";
import { TextField } from "#/components/ui/TextField/TextField";
import { toast } from "#/components/ui/ToastRegion/toast";
import { ImagePurpose } from "#/gen/chat/v1/image_service_pb";

import { useWebhookActions } from "../hooks/useWebhooks";
import { webhookUrl } from "../utils/webhookUrl";
import { WebhookUrlPanel } from "./WebhookUrlPanel";

import type { Webhook } from "#/gen/chat/v1/webhook_service_pb";

type WebhookDialogProps = {
  channelId: string;
  // null なら新規発行、あればその Webhook を編集する
  webhook: Webhook | null;
  onClose: () => void;
};

export const WebhookDialog = ({ channelId, webhook, onClose }: WebhookDialogProps) => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ strict: false });
  const [name, setName] = useState(webhook?.name ?? "");
  const [avatarUrl, setAvatarUrl] = useState(webhook?.avatarUrl ?? "");
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  // 発行・再発行した直後だけ URL を表示する
  const [revealedUrl, setRevealedUrl] = useState<string | null>(null);
  const { create, regenerate, remove, update } = useWebhookActions(channelId);
  const trimmedName = name.trim();
  const failed = [create, update, regenerate, remove].find((mutation) => mutation.isError);

  const save = () => {
    setIsSubmitted(true);
    if (trimmedName === "") {
      return;
    }
    const input = { avatarUrl: avatarUrl || undefined, name: trimmedName };
    if (webhook) {
      update.mutate(
        { ...input, webhookId: webhook.id },
        {
          onSuccess: () => {
            toast(t("webhook.updated"));
            onClose();
          },
        },
      );
      return;
    }
    create.mutate(
      { ...input, channelId },
      {
        onSuccess: (res) => {
          toast(t("webhook.created"));
          setRevealedUrl(webhookUrl(res.webhook?.id ?? "", res.token));
        },
      },
    );
  };

  const onOpenChange = (isOpen: boolean) => {
    if (!isOpen) {
      onClose();
    }
  };

  if (revealedUrl !== null) {
    return (
      <Dialog
        isOpen
        onOpenChange={onOpenChange}
        title={webhook ? t("webhook.editTitle") : t("webhook.createTitle")}
        size="md"
        footer={<Button onPress={onClose}>{t("webhook.done")}</Button>}
      >
        <WebhookUrlPanel url={revealedUrl} />
      </Dialog>
    );
  }

  return (
    <Dialog
      isOpen
      onOpenChange={onOpenChange}
      title={webhook ? t("webhook.editTitle") : t("webhook.createTitle")}
      footer={
        <>
          {webhook && (
            <Button
              variant="ghost"
              className="mr-auto text-danger"
              onPress={() => {
                setIsDeleteOpen(true);
              }}
            >
              {t("common.delete")}
            </Button>
          )}
          <Button variant="secondary" onPress={onClose}>
            {t("common.cancel")}
          </Button>
          <Button isPending={create.isPending || update.isPending} onPress={save}>
            {webhook ? t("common.save") : t("webhook.submit")}
          </Button>
        </>
      }
    >
      <p className="m-0 text-caption text-muted">{t("webhook.hint")}</p>
      <TextField
        label={t("webhook.name")}
        placeholder={t("webhook.namePlaceholder")}
        value={name}
        onChange={setName}
        isRequired
        maxLength={80}
        errorMessage={isSubmitted && trimmedName === "" ? t("webhook.nameRequired") : undefined}
      />
      <IconImageField
        label={t("webhook.avatar")}
        name={trimmedName || t("webhook.namePlaceholder")}
        value={avatarUrl}
        onChange={setAvatarUrl}
        purpose={ImagePurpose.WEBHOOK_ICON}
        workspaceId={workspaceId ?? null}
      />
      {webhook && (
        <Button
          variant="secondary"
          size="sm"
          className="self-start"
          isPending={regenerate.isPending}
          onPress={() => {
            regenerate.mutate(
              { webhookId: webhook.id },
              {
                onSuccess: (res) => {
                  toast(t("webhook.regenerated"));
                  setRevealedUrl(webhookUrl(webhook.id, res.token));
                },
              },
            );
          }}
        >
          {t("webhook.regenerate")}
        </Button>
      )}
      {failed && <p className="m-0 text-caption text-danger">{failed.error.message}</p>}
      {webhook && (
        <AlertDialog
          isOpen={isDeleteOpen}
          onOpenChange={setIsDeleteOpen}
          title={t("webhook.delete.title", { name: webhook.name })}
          confirmLabel={t("webhook.delete.confirm")}
          tone="danger"
          isPending={remove.isPending}
          onConfirm={() => {
            remove.mutate(
              { webhookId: webhook.id },
              {
                onSuccess: () => {
                  toast(t("webhook.delete.done"));
                  onClose();
                },
              },
            );
          }}
        >
          {t("webhook.delete.body")}
        </AlertDialog>
      )}
    </Dialog>
  );
};
