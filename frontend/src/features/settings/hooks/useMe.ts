import { useQuery } from "@tanstack/react-query";

import { api } from "#/lib/api/client";

/** ログイン中ユーザーのプロフィールを取得する */
export const useMe = () =>
  useQuery({
    queryFn: async () => {
      const { data, error } = await api.GET("/api/users/me");
      if (error) {
        throw new Error(error.error);
      }
      return data;
    },
    queryKey: ["users", "me"],
  });
