import { useState } from "react";

import { Button, Checkbox, Group, Stack, Text, TextInput, Textarea } from "@mantine/core";

import { useUpdateChannel } from "../hooks/useUpdateChannel";

type Props = {
  channelId: string;
  initialName: string;
  initialDescription: string | null;
  initialIsPrivate: boolean;
};

export const ChannelSettingsPanel = ({ channelId, initialName, initialDescription, initialIsPrivate }: Props) => {
  const update = useUpdateChannel();
  const [name, setName] = useState<string>(initialName);
  const [description, setDescription] = useState<string>(initialDescription ?? "");
  const [isPrivate, setIsPrivate] = useState<boolean>(initialIsPrivate);

  const onSubmit = async () => {
    await update.mutateAsync({ channelId, name, description, isPrivate });
  };

  return (
    <Stack gap="sm">
      <Text fw={600}>チャンネル設定</Text>
      <TextInput label="名前" value={name} onChange={(e) => setName(e.currentTarget.value)} required />
      <Textarea label="説明" value={description} onChange={(e) => setDescription(e.currentTarget.value)} autosize minRows={3} />
      <Checkbox label="プライベートチャンネル" checked={isPrivate} onChange={(e) => setIsPrivate(e.currentTarget.checked)} />
      <Group justify="flex-end">
        <Button onClick={onSubmit} loading={update.isPending}>
          保存
        </Button>
      </Group>
      {update.isError && (
        <Text c="red" size="sm">{(update.error as Error)?.message ?? "更新に失敗しました"}</Text>
      )}
      {update.isSuccess && <Text c="green" size="sm">保存しました</Text>}
    </Stack>
  );
};


