import { useState } from "react";

import { Modal, TextInput, Textarea, Button, Group } from "@mantine/core";
import { useForm } from "@mantine/form";

import { useUserGroups } from "../hooks/useUserGroups";

import type { CreateUserGroupInput } from "../types";

type CreateUserGroupModalProps = {
  opened: boolean;
  onClose: () => void;
  workspaceId: string;
  onSuccess: () => void;
};

export const CreateUserGroupModal = ({
  opened,
  onClose,
  workspaceId,
  onSuccess,
}: CreateUserGroupModalProps) => {
  const { createGroup } = useUserGroups(workspaceId);
  const [isLoading, setIsLoading] = useState(false);

  const form = useForm<CreateUserGroupInput>({
    initialValues: {
      workspaceId,
      name: "",
      description: "",
    },
    validate: {
      name: (value) => {
        if (!value.trim()) return "グループ名は必須です";
        if (value.length < 2) return "グループ名は2文字以上にしてください";
        if (value.length > 50) return "グループ名は50文字以内にしてください";
        if (!/^[a-zA-Z0-9_-]+$/.test(value)) {
          return "グループ名は英数字、ハイフン、アンダースコアのみ使用できます";
        }
        return null;
      },
      description: (value) => {
        if (value && value.length > 200) {
          return "説明は200文字以内にしてください";
        }
        return null;
      },
    },
  });

  const handleSubmit = async (values: CreateUserGroupInput) => {
    setIsLoading(true);
    try {
      await createGroup(values);
      form.reset();
      onSuccess();
    } catch (error) {
      console.error("グループの作成に失敗しました:", error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleClose = () => {
    form.reset();
    onClose();
  };

  return (
    <Modal opened={opened} onClose={handleClose} title="ユーザーグループの作成" size="md">
      <form onSubmit={form.onSubmit(handleSubmit)}>
        <div className="space-y-4">
          <TextInput
            label="グループ名"
            placeholder="例: 開発者, マーケティング"
            description="これは @ メンションで使用されます (例: @developers)"
            required
            {...form.getInputProps("name")}
          />

          <Textarea
            label="説明"
            placeholder="グループの説明を入力"
            rows={3}
            {...form.getInputProps("description")}
          />

          <Group justify="flex-end" mt="md">
            <Button variant="subtle" onClick={handleClose}>
              キャンセル
            </Button>
            <Button type="submit" loading={isLoading}>
              作成
            </Button>
          </Group>
        </div>
      </form>
    </Modal>
  );
};
