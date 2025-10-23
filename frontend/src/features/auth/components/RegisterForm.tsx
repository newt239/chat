import { useState } from "react";
import { TextInput, PasswordInput, Button, Paper, Title, Text, Anchor } from "@mantine/core";
import { useRegister } from "../hooks/useAuth";

export function RegisterForm() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [displayName, setDisplayName] = useState("");
  const register = useRegister();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    register.mutate({ email, password, displayName });
  };

  return (
    <Paper className="w-full max-w-md p-8" shadow="md" radius="md">
      <Title order={2} className="mb-6 text-center">
        新規登録
      </Title>

      <form onSubmit={handleSubmit}>
        <TextInput
          label="表示名"
          placeholder="山田太郎"
          value={displayName}
          onChange={(e) => setDisplayName(e.currentTarget.value)}
          required
          className="mb-4"
        />

        <TextInput
          label="メールアドレス"
          placeholder="your@email.com"
          type="email"
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
          <Anchor href="/login" size="sm">
            ログイン
          </Anchor>
        </Text>
      </form>
    </Paper>
  );
}
