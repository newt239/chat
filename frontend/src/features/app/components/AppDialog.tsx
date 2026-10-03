import { useState } from "react";

import { useTranslation } from "react-i18next";

import { CopyableUrl } from "#/components/block/CopyableUrl/CopyableUrl";
import { IconImageField } from "#/components/block/IconImageField/IconImageField";
import { AlertDialog } from "#/components/ui/AlertDialog/AlertDialog";
import { Button } from "#/components/ui/Button/Button";
import { Checkbox } from "#/components/ui/Checkbox/Checkbox";
import { Dialog } from "#/components/ui/Dialog/Dialog";
import { Select } from "#/components/ui/Select/Select";
import { TextField } from "#/components/ui/TextField/TextField";
import { toast } from "#/components/ui/ToastRegion/toast";
import { useChannels } from "#/features/channel/hooks/useChannel";
import { AppPermission } from "#/gen/chat/v1/app_service_pb";
import { ImagePurpose } from "#/gen/chat/v1/image_service_pb";
import { apiBaseUrl } from "#/lib/api/createTransport";

import { useAppActions } from "../hooks/useApps";

import type { App } from "#/gen/chat/v1/app_service_pb";

type AppDialogProps = {
  workspaceId: string;
  // null なら新規作成、あればそのアプリを編集する
  app: App | null;
  // 新規作成のときの既定の投稿先。チャンネルから開いたときはそのチャンネル
  initialChannelId: string | null;
  onClose: () => void;
};

const NO_CHANNEL = "";

// 画面に並べる順と、辞書の app.permissions.* のキー
const appPermissionKeys = [
  [AppPermission.POST_JOINED_CHANNELS, "postJoinedChannels"],
  [AppPermission.POST_PUBLIC_CHANNELS, "postPublicChannels"],
  [AppPermission.POST_THREAD_REPLIES, "postThreadReplies"],
  [AppPermission.OUTGOING_WEBHOOK, "outgoingWebhook"],
] as const;

const appWebhookUrl = (appId: string, token: string) =>
  new URL(`/webhooks/${appId}/${token}`, apiBaseUrl).href;

export const AppDialog = ({ workspaceId, app, initialChannelId, onClose }: AppDialogProps) => {
  const { t } = useTranslation();
  const { data: channels = [] } = useChannels(workspaceId);
  const [name, setName] = useState(app?.name ?? "");
  const [description, setDescription] = useState(app?.description ?? "");
  const [avatarUrl, setAvatarUrl] = useState(app?.avatarUrl ?? "");
  const [permissions, setPermissions] = useState<AppPermission[]>(
    app?.permissions ?? [AppPermission.POST_JOINED_CHANNELS],
  );
  const [defaultChannelId, setDefaultChannelId] = useState(
    app ? (app.defaultChannelId ?? NO_CHANNEL) : (initialChannelId ?? NO_CHANNEL),
  );
  const [outgoingUrl, setOutgoingUrl] = useState(app?.outgoingUrl ?? "");
  const [isSubmitted, setIsSubmitted] = useState(false);
  const [isDeleteOpen, setIsDeleteOpen] = useState(false);
  // 発行・再発行した直後だけ URL を表示する。サーバーはトークンをハッシュでしか持たないため再表示できない
  const [revealedUrl, setRevealedUrl] = useState<string | null>(null);
  const { create, regenerate, remove, update } = useAppActions();
  const trimmedName = name.trim();
  const hasOutgoing = permissions.includes(AppPermission.OUTGOING_WEBHOOK);
  const isOutgoingMissing = hasOutgoing && outgoingUrl.trim() === "";
  const failed = [create, update, regenerate, remove].find((mutation) => mutation.isError);
  // DM には参加できないため、投稿先には公開・非公開チャンネルだけを出す
  const channelOptions = [
    { label: t("app.defaultChannelNone"), value: NO_CHANNEL },
    ...channels.map((channel) => ({ label: `#${channel.name}`, value: channel.id })),
  ];

  const togglePermission = (permission: AppPermission, isSelected: boolean) => {
    setPermissions((current) =>
      isSelected ? [...current, permission] : current.filter((value) => value !== permission),
    );
  };

  const save = () => {
    setIsSubmitted(true);
    if (trimmedName === "" || isOutgoingMissing) {
      return;
    }
    const settings = {
      avatarUrl: avatarUrl || undefined,
      defaultChannelId: defaultChannelId || undefined,
      description: description.trim() || undefined,
      name: trimmedName,
      outgoingUrl: hasOutgoing ? outgoingUrl.trim() : undefined,
      permissions,
    };
    if (app) {
      update.mutate(
        { appId: app.id, settings },
        {
          onSuccess: () => {
            toast(t("app.updated"));
            onClose();
          },
        },
      );
      return;
    }
    create.mutate(
      { settings, workspaceId },
      {
        onSuccess: (res) => {
          toast(t("app.created"));
          setRevealedUrl(appWebhookUrl(res.app?.id ?? "", res.token));
        },
      },
    );
  };

  const title = app ? t("app.editTitle") : t("app.createTitle");

  if (revealedUrl !== null) {
    return (
      <Dialog
        isOpen
        onOpenChange={onClose}
        title={title}
        size="md"
        footer={<Button onPress={onClose}>{t("app.done")}</Button>}
      >
        <p className="m-0 rounded-md bg-accent-soft px-3 py-2 text-caption text-accent-text">
          {t("app.urlOnce")}
        </p>
        <span className="text-xs font-semibold text-muted">{t("app.url")}</span>
        <CopyableUrl url={revealedUrl} />
        <span className="text-xs font-semibold text-muted">{t("app.example")}</span>
        <pre className="m-0 overflow-x-auto rounded-md border border-border bg-sunken px-2 py-1.5 font-mono text-xs whitespace-pre">
          {`curl -X POST -H 'Content-Type: application/json' \\\n  -d '{"text": "Hello", "channel_id": "<任意>", "thread_id": "<任意>"}' \\\n  ${revealedUrl}`}
        </pre>
      </Dialog>
    );
  }

  return (
    <Dialog
      isOpen
      onOpenChange={onClose}
      title={title}
      size="md"
      footer={
        <>
          {app && (
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
            {app ? t("common.save") : t("app.submit")}
          </Button>
        </>
      }
    >
      <p className="m-0 text-caption text-muted">{t("app.hint")}</p>
      <TextField
        label={t("app.name")}
        placeholder={t("app.namePlaceholder")}
        value={name}
        onChange={setName}
        isRequired
        maxLength={80}
        errorMessage={isSubmitted && trimmedName === "" ? t("app.nameRequired") : undefined}
      />
      <TextField
        label={t("app.description")}
        value={description}
        onChange={setDescription}
        maxLength={500}
      />
      <IconImageField
        label={t("app.avatar")}
        name={trimmedName || t("app.namePlaceholder")}
        value={avatarUrl}
        onChange={setAvatarUrl}
        purpose={ImagePurpose.APP_ICON}
        workspaceId={workspaceId}
      />
      <fieldset className="m-0 flex flex-col gap-1.5 border-0 p-0">
        <legend className="mb-1 text-xs font-semibold text-muted">
          {t("app.permissions.label")}
        </legend>
        {appPermissionKeys.map(([permission, key]) => (
          <Checkbox
            key={key}
            isSelected={permissions.includes(permission)}
            onChange={(isSelected) => {
              togglePermission(permission, isSelected);
            }}
          >
            <span className="flex flex-col">
              {t(`app.permissions.${key}`)}
              <small className="text-caption text-muted">{t(`app.permissions.${key}Hint`)}</small>
            </span>
          </Checkbox>
        ))}
      </fieldset>
      {hasOutgoing && (
        <TextField
          label={t("app.outgoingUrl")}
          placeholder="https://example.com/webhook"
          type="url"
          value={outgoingUrl}
          onChange={setOutgoingUrl}
          isRequired
          errorMessage={isSubmitted && isOutgoingMissing ? t("app.outgoingUrlRequired") : undefined}
        />
      )}
      {app?.outgoingSecret && (
        <div className="flex flex-col gap-1">
          <span className="text-xs font-semibold text-muted">{t("app.outgoingSecret")}</span>
          <CopyableUrl url={app.outgoingSecret} />
          <small className="text-caption text-muted">{t("app.outgoingSecretHint")}</small>
        </div>
      )}
      <Select
        label={t("app.defaultChannel")}
        description={t("app.defaultChannelHint")}
        options={channelOptions}
        value={defaultChannelId}
        onChange={setDefaultChannelId}
      />
      {app && (
        <Button
          variant="secondary"
          size="sm"
          className="self-start"
          isPending={regenerate.isPending}
          onPress={() => {
            regenerate.mutate(
              { appId: app.id },
              {
                onSuccess: (res) => {
                  toast(t("app.regenerated"));
                  setRevealedUrl(appWebhookUrl(app.id, res.token));
                },
              },
            );
          }}
        >
          {t("app.regenerate")}
        </Button>
      )}
      {failed && <p className="m-0 text-caption text-danger">{failed.error.message}</p>}
      {app && (
        <AlertDialog
          isOpen={isDeleteOpen}
          onOpenChange={setIsDeleteOpen}
          title={t("app.delete.title", { name: app.name })}
          confirmLabel={t("app.delete.confirm")}
          tone="danger"
          isPending={remove.isPending}
          onConfirm={() => {
            remove.mutate(
              { appId: app.id },
              {
                onSuccess: () => {
                  toast(t("app.delete.done"));
                  onClose();
                },
              },
            );
          }}
        >
          {t("app.delete.body")}
        </AlertDialog>
      )}
    </Dialog>
  );
};
