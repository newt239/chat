import { useMutation } from "@connectrpc/connect-query";

import { useCompleteLogin } from "#/features/auth/hooks/useCompleteLogin";
import { AuthService } from "#/gen/chat/v1/auth_service_pb";

export const useLogin = () => {
  const completeLogin = useCompleteLogin();
  return useMutation(AuthService.method.login, { onSuccess: completeLogin });
};
