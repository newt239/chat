import { useMutation } from "@connectrpc/connect-query";

import { useCompleteLogin } from "#/features/auth/hooks/useCompleteLogin";
import { AuthService } from "#/gen/chat/v1/auth_service_pb";

export const useSignUp = (workspaceId: string) => {
  const completeLogin = useCompleteLogin(workspaceId);
  return useMutation(AuthService.method.signUp, { onSuccess: completeLogin });
};
