import { useMutation } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { api } from "#/lib/api/client";
import { setAuthAtom } from "#/providers/store/auth";

import type { components } from "#/lib/api/schema";

type AuthResponse = components["schemas"]["AuthResponse"];

export const useRegister = () => {
  const setAuth = useSetAtom(setAuthAtom);
  const navigate = useNavigate();

  return useMutation({
    mutationFn: async (data: { email: string; password: string; displayName: string }) => {
      const { data: response, error } = await api.POST("/api/auth/register", {
        body: data,
      });
      if (error) {
        throw new Error(error.error);
      }
      return response;
    },
    onSuccess: (data: AuthResponse) => {
      setAuth({ accessToken: data.accessToken, refreshToken: data.refreshToken, user: data.user });
      void navigate({ to: "/app" });
    },
  });
};
