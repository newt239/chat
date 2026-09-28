import { useState } from "react";

import { useParams } from "@tanstack/react-router";
import { Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button";
import { Checkbox } from "#/components/ui/Checkbox";
import { TextArea } from "#/components/ui/TextArea";
import { toast } from "#/components/ui/toast";
import {
  channelPathErrorKeys,
  lastSegment,
  parentPath,
  validateChannelPath,
} from "#/features/channel/utils/channelPath";

import { useUpdateChannel } from "../hooks/useUpdateChannel";
import { ChannelNameField } from "./ChannelNameField";

type Props = {
  channelId: string;
  initialName: string;
  initialDescription: string | null;
  initialIsPrivate: boolean;
};

// 名前はパスの末尾だけを編集させ、親のパスは固定の接頭辞として見せる
export const ChannelSettingsPanel = ({
  channelId,
  initialName,
  initialDescription,
  initialIsPrivate,
}: Props) => {
  const { t } = useTranslation();
  const { workspaceId } = useParams({ from: "/app/$workspaceId" });
  const update = useUpdateChannel(workspaceId);
  const parent = parentPath(initialName);
  const [segment, setSegment] = useState(lastSegment(initialName));
  const [description, setDescription] = useState(initialDescription ?? "");
  const [isPrivate, setIsPrivate] = useState(initialIsPrivate);

  const nameError = segment.includes("/") ? "invalid" : validateChannelPath(segment, []);
  const name = parent === null ? segment : `${parent}/${segment}`;

  return (
    <Form
      className="flex flex-col gap-3 px-4 py-3"
      onSubmit={(event) => {
        event.preventDefault();
        if (nameError !== null) {
          return;
        }
        update.mutate(
          { channelId, description, isPrivate, name },
          {
            onSuccess: () => {
              toast(t("channel.settings.saved"), { tone: "success" });
            },
          },
        );
      }}
    >
      <h4 className="m-0 text-xs font-semibold text-muted">{t("channel.settings.title")}</h4>
      <ChannelNameField
        label={t("channel.settings.name")}
        prefix={parent === null ? "#" : `#${parent}/`}
        value={segment}
        onChange={setSegment}
        description={t("channel.settings.nameHint")}
        errorMessage={nameError === null ? null : t(channelPathErrorKeys[nameError], { name })}
      />
      <TextArea
        label={t("channel.settings.description")}
        value={description}
        onChange={setDescription}
      />
      <Checkbox isSelected={isPrivate} onChange={setIsPrivate}>
        {t("channel.settings.private")}
      </Checkbox>
      <div className="flex justify-end">
        <Button type="submit" isPending={update.isPending}>
          {t("common.save")}
        </Button>
      </div>
      {update.isError && (
        <p role="alert" className="m-0 text-caption text-danger">
          {update.error.message}
        </p>
      )}
    </Form>
  );
};
