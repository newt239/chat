import { useEffect, useId, useState } from "react";

import { IconChevronRight } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { Dialog } from "#/components/ui/Dialog/Dialog";
import { Switch } from "#/components/ui/Switch/Switch";
import { TextArea } from "#/components/ui/TextArea/TextArea";
import { toast } from "#/components/ui/ToastRegion/toast";
import { useChannels, useCreateChannel } from "#/features/channel/hooks/useChannel";
import {
  ancestorPaths,
  channelPathErrorKeys,
  validateChannelPath,
} from "#/features/channel/utils/channelPath";

import { ChannelNameField } from "./ChannelNameField";

type CreateChannelModalProps = {
  workspaceId: string;
  // 子チャンネルとして作るときの親
  parentId: string | null;
  onClose: () => void;
};

// 開くたびにマウントし直すため、入力は閉じるときに戻さなくてよい
export const CreateChannelModal = ({ workspaceId, parentId, onClose }: CreateChannelModalProps) => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const formId = useId();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [isPrivate, setIsPrivate] = useState(false);
  const [isTouched, setIsTouched] = useState(false);
  const { data: channels } = useChannels(workspaceId);
  const createChannel = useCreateChannel();
  const parent = channels?.find((channel) => channel.id === parentId);
  const parentName = parent?.name;
  const isParentPrivate = parent?.isPrivate ?? false;

  // 親を指定して開いたら、一覧が読み込まれた時点で親のパスと公開範囲を初期値にする
  useEffect(() => {
    if (parentName !== undefined) {
      setName(`${parentName}/`);
      setIsPrivate(isParentPrivate);
    }
  }, [parentName, isParentPrivate]);

  const existingNames = (channels ?? []).map((channel) => channel.name);
  const error = validateChannelPath(name, existingNames);
  const segments = name.split("/").filter((segment) => segment.length > 0);
  const missingParents = ancestorPaths(name).filter((path) => !existingNames.includes(path));

  const submit = () => {
    setIsTouched(true);
    if (error !== null) {
      return;
    }
    createChannel.mutate(
      { description: description || undefined, isPrivate, name, workspaceId },
      {
        onSuccess: ({ channel }) => {
          toast(t("channel.create.created", { name }), { tone: "success" });
          // 移動先に ?dialog= がないため、移動するとダイアログも閉じる
          if (channel === undefined) {
            onClose();
          } else {
            void navigate({
              params: { channelId: channel.id, workspaceId },
              to: "/app/$workspaceId/$channelId",
            });
          }
        },
      },
    );
  };

  return (
    <Dialog
      isOpen
      onOpenChange={(isOpen) => {
        if (!isOpen) {
          onClose();
        }
      }}
      title={t("channel.create.title")}
      footer={
        <>
          <Button variant="secondary" onPress={onClose}>
            {t("common.cancel")}
          </Button>
          <Button type="submit" form={formId} isPending={createChannel.isPending}>
            {t("channel.create.submit")}
          </Button>
        </>
      }
    >
      <Form
        id={formId}
        className="flex flex-col gap-3"
        onSubmit={(event) => {
          event.preventDefault();
          submit();
        }}
      >
        <ChannelNameField
          label={t("channel.create.name")}
          prefix="#"
          value={name}
          onChange={setName}
          placeholder="dev/frontend"
          description={t("channel.create.nameHint")}
          errorMessage={
            isTouched && error !== null ? t(channelPathErrorKeys[error], { name }) : null
          }
        />
        {segments.length > 1 && (
          <ol
            aria-label={t("channel.create.preview")}
            className="m-0 flex list-none flex-wrap items-center gap-1 rounded-md bg-sunken px-2.5 py-2 font-mono text-xs text-muted"
          >
            {segments.map((segment, index) => (
              <li key={segments.slice(0, index + 1).join("/")} className="flex items-center gap-1">
                {index > 0 && <IconChevronRight aria-hidden className="size-3" />}
                {index === segments.length - 1 ? (
                  <b className="font-medium text-text"># {segment}</b>
                ) : (
                  <span># {segment}</span>
                )}
              </li>
            ))}
          </ol>
        )}
        {error === null && missingParents.length > 0 && (
          <p className="m-0 text-xs text-muted">
            {t("channel.create.parentsCreated", {
              names: missingParents.map((path) => `#${path}`).join(", "),
            })}
          </p>
        )}
        <TextArea
          label={t("channel.create.description")}
          placeholder={t("channel.create.descriptionPlaceholder")}
          value={description}
          onChange={setDescription}
          rows={2}
        />
        <div className="flex flex-col gap-0.5">
          <Switch isSelected={isPrivate} onChange={setIsPrivate}>
            {t("channel.create.private")}
          </Switch>
          <span className="pl-11 text-caption text-muted">{t("channel.create.privateHint")}</span>
        </div>
        {createChannel.isError && (
          <p role="alert" className="m-0 text-caption text-danger">
            {createChannel.error.message}
          </p>
        )}
      </Form>
    </Dialog>
  );
};
