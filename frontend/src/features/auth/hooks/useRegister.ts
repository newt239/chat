import { useMutation } from "@connectrpc/connect-query";
import { useNavigate } from "@tanstack/react-router";
import { useSetAtom } from "jotai";

import { AuthService } from "#/gen/chat/v1/auth_service_pb";
import { setAuthAtom } from "#/providers/store/auth";

export const useRegister = () => {
  const setAuth = useSetAtom(setAuthAtom);
  const navigate = useNavigate();

  return useMutation(AuthService.method.register, {
    onSuccess: ({ accessToken, refreshToken, user }) => {
      setAuth({ accessToken, refreshToken, user });
      void navigate({ to: "/app" });
    },
  });
};
