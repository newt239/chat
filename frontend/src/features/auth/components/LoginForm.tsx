import { useState } from "react";

import { TextInput, PasswordInput, Button, Paper, Title, Text, Anchor } from "@mantine/core";

import { useLogin } from "../hooks/useAuth";

export const LoginForm = () => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const login = useLogin();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    login.mutate({ email, password });
  };

  return (
    <Paper className="w-full max-w-md p-8" shadow="md" radius="md">
      <Title order={2} className="mb-6 text-center">
        ログイン
      </Title>

      <form onSubmit={handleSubmit}>
        <TextInput
          label="メールアドレス"
          placeholder="your@email.com"
          value={email}
          onChange={(e) => setEmail(e.currentTarget.value)}
          required
          className="mb-4"
        />

        <PasswordInput
          label="パスワード"
          placeholder="パスワード"
          value={password}
          onChange={(e) => setPassword(e.currentTarget.value)}
          required
          className="mb-6"
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
          <Anchor href="/register" size="sm">
            新規登録
          </Anchor>
        </Text>
      </form>
    </Paper>
  );
}
