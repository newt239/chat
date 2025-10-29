import { Anchor, Button, Paper, PasswordInput, Text, TextInput, Title } from "@mantine/core";
import { useForm } from "@mantine/form";
import { Link } from "@tanstack/react-router";

import { useRegister } from "@/features/auth/hooks/useRegister";

type RegisterFormValues = {
  displayName: string;
  email: string;
  password: string;
};

export const RegisterForm = () => {
  const register = useRegister();

  const form = useForm<RegisterFormValues>({
    initialValues: {
      displayName: "",
      email: "",
      password: "",
    },
    validate: {
      displayName: (value) => (value.length >= 1 ? null : "1文字以上の表示名を入力してください"),
      email: (value) => (/^\S+@\S+$/.test(value) ? null : "有効なメールアドレスを入力してください"),
      password: (value) => (value.length >= 8 ? null : "8文字以上のパスワードを入力してください"),
    },
  });

  const handleSubmit = form.onSubmit((values) => {
    register.mutate(values);
  });

  return (
    <Paper className="w-full max-w-md p-8" shadow="md" radius="md">
      <Title order={2} className="mb-6 text-center">
        新規登録
      </Title>

      <form onSubmit={handleSubmit}>
        <TextInput
          label="表示名"
          placeholder="newt"
          {...form.getInputProps("displayName")}
          required
          className="mb-4"
        />

        <TextInput
          label="メールアドレス"
          placeholder="email@example.com"
          type="email"
          {...form.getInputProps("email")}
          required
          className="mb-4"
        />

        <PasswordInput
          label="パスワード"
          placeholder="パスワード"
          {...form.getInputProps("password")}
          required
          className="mb-6"
        />

        {register.isError && (
          <Text c="red" size="sm" className="mb-4">
            {register.error?.message || "登録に失敗しました"}
          </Text>
        )}

        <Button type="submit" fullWidth loading={register.isPending} className="mb-4">
          登録
        </Button>

        <Text size="sm" className="text-center">
          すでにアカウントをお持ちの方は{" "}
          <Anchor component={Link} to="/login" size="sm">
            ログイン
          </Anchor>
        </Text>
      </form>
    </Paper>
  );
};
