import { useMutation } from "@tanstack/react-query";

import { store } from "@/providers/store";
import { authAtom } from "@/providers/store/auth";

type UpdateProfileInput = {
  displayName?: string;
  bio?: string | null;
  avatarUrl?: string | null;
};

type UpdateMeResponse = {
  id: string;
  displayName: string;
  bio?: string | null;
  avatarURL?: string | null;
};

export function useUpdateProfile() {
  return useMutation({
    mutationFn: async (input: UpdateProfileInput) => {
      const baseUrl = import.meta.env.VITE_API_BASE_URL || "";

      const accessToken = store.get(authAtom).accessToken;
      const headers = new Headers({ "Content-Type": "application/json" });
      if (accessToken) headers.set("Authorization", `Bearer ${accessToken}`);

      const body: Record<string, unknown> = {};
      if (input.displayName !== undefined) body.display_name = input.displayName;
      if (input.bio !== undefined) body.bio = input.bio;
      if (input.avatarUrl !== undefined) body.avatar_url = input.avatarUrl;

      const res = await fetch(`${baseUrl}/api/users/me`, {
        method: "PATCH",
        headers,
        body: JSON.stringify(body),
      });
      if (!res.ok) {
        const text = await res.text();
        throw new Error(text || "プロフィール更新に失敗しました");
      }
      const data = (await res.json()) as UpdateMeResponse;

      // auth の user を部分更新（型上存在するフィールドのみ反映）
      const current = store.get(authAtom);
      const nextUser = current.user
        ? { ...current.user, displayName: data.displayName, avatarUrl: data.avatarURL ?? null }
        : current.user;
      store.set(authAtom, {
        user: nextUser,
        accessToken: current.accessToken,
        refreshToken: current.refreshToken,
      });

      return data;
    },
  });
}


