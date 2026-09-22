import { useMutation } from "@tanstack/react-query";

import { api } from "#/lib/api/client";
import { store } from "#/providers/store";
import { authAtom } from "#/providers/store/auth";

type UpdateProfileInput = {
  displayName?: string;
  bio?: string | null;
  avatarUrl?: string | null;
};

export const useUpdateProfile = () =>
  useMutation({
    mutationFn: async (input: UpdateProfileInput) => {
      const { data, error } = await api.PATCH("/api/users/me", {
        body: {
          avatar_url: input.avatarUrl,
          bio: input.bio,
          display_name: input.displayName,
        },
      });
      if (error) {
        throw new Error(error.error);
      }

      // Auth の user を部分更新（型上存在するフィールドのみ反映）
      const current = store.get(authAtom);
      store.set(authAtom, {
        accessToken: current.accessToken,
        refreshToken: current.refreshToken,
        user: current.user
          ? { ...current.user, avatarUrl: data.avatarUrl ?? null, displayName: data.displayName }
          : current.user,
      });

      return data;
    },
  });
