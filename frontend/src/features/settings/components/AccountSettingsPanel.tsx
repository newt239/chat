import { useState } from "react";

import { Button, Group, PasswordInput, Stack, Text } from "@mantine/core";

import { useDeleteAccount, useUpdatePassword } from "#/features/settings/hooks/useAccount";

export const AccountSettingsPanel = () => {
  const updatePassword = useUpdatePassword();
  const deleteAccount = useDeleteAccount();

  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [isDeleteConfirming, setIsDeleteConfirming] = useState(false);

  const canSubmit = currentPassword.length > 0 && newPassword.length >= 8;

  const handleUpdatePassword = () => {
    updatePassword.mutate(
      { currentPassword, newPassword },
      {
        onSuccess: () => {
          setCurrentPassword("");
          setNewPassword("");
        },
      },
    );
  };

  return (
    <Stack gap="sm">
      <Text fw={600}>パスワードの変更</Text>
      <PasswordInput
        label="現在のパスワード"
        value={currentPassword}
        onChange={(event) => {
          setCurrentPassword(event.currentTarget.value);
        }}
      />
      <PasswordInput
        label="新しいパスワード"
        description="8文字以上"
        value={newPassword}
        onChange={(event) => {
          setNewPassword(event.currentTarget.value);
        }}
      />
      <Group justify="flex-end">
        <Button
          disabled={!canSubmit}
          loading={updatePassword.isPending}
          onClick={handleUpdatePassword}
        >
          変更
        </Button>
      </Group>
      {updatePassword.isError && (
        <Text c="red" size="sm">
          {updatePassword.error.message}
        </Text>
      )}
      {updatePassword.isSuccess && (
        <Text c="green" size="sm">
          パスワードを変更しました。再度ログインしてください
        </Text>
      )}

      <Text fw={600} c="red">
        アカウントの削除
      </Text>
      <Text size="xs" c="dimmed">
        投稿したメッセージを含め、アカウントに紐づくデータが削除されます。この操作は取り消せません。
      </Text>
      {isDeleteConfirming ? (
        <Group justify="flex-end">
          <Button
            variant="subtle"
            onClick={() => {
              setIsDeleteConfirming(false);
            }}
          >
            キャンセル
          </Button>
          <Button
            color="red"
            loading={deleteAccount.isPending}
            onClick={() => {
              deleteAccount.mutate();
            }}
          >
            削除する
          </Button>
        </Group>
      ) : (
        <Group justify="flex-end">
          <Button
            variant="outline"
            color="red"
            onClick={() => {
              setIsDeleteConfirming(true);
            }}
          >
            アカウントを削除
          </Button>
        </Group>
      )}
      {deleteAccount.isError && (
        <Text c="red" size="sm">
          {deleteAccount.error.message}
        </Text>
      )}
    </Stack>
  );
};
