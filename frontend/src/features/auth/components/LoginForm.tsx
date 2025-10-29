import { Anchor, Button, Paper, PasswordInput, Text, TextInput, Title } from "@mantine/core";
import { useForm } from "@mantine/form";
import { Link } from "@tanstack/react-router";

import { useLogin } from "../hooks/useAuth";

type LoginFormValues = {
  email: string;
  password: string;
};

export const LoginForm = () => {
  const login = useLogin();

  const form = useForm<LoginFormValues>({
    initialValues: {
      email: "",
      password: "",
    },
    validate: {
      email: (value) => (/^\S+@\S+$/.test(value) ? null : "有効なメールアドレスを入力してください"),
      password: (value) => (value.length >= 6 ? null : "6文字以上のパスワードを入力してください"),
    },
  });

  const handleSubmit = form.onSubmit((values) => {
    login.mutate(values);
  });

  return (
    <Paper className="w-full max-w-md p-8" shadow="md" radius="md">
      <Title order={2} className="mb-6 text-center">
        ログイン
      </Title>

      <form onSubmit={handleSubmit}>
        <TextInput
          label="メールアドレス"
          placeholder="email@example.com"
          type="email"
          required
          className="mb-4"
          {...form.getInputProps("email")}
        />

        <PasswordInput
          label="パスワード"
          placeholder="パスワード"
          required
          className="mb-6"
          {...form.getInputProps("password")}
        />

        {login.isError && (
          <Text c="red" size="sm" className="mb-4">
            {login.error?.message || "ログインに失敗しました"}
          </Text>
        )}

        <Button type="submit" fullWidth loading={login.isPending} className="mb-4">
          ログイン
        </Button>

        <Text size="sm" className="text-center">
          アカウントをお持ちでない方は{" "}
          <Anchor component={Link} to="/register" size="sm">
            新規登録
          </Anchor>
        </Text>
      </form>
    </Paper>
  );
};
