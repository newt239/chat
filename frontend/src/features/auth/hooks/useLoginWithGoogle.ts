import { useMutation } from "@connectrpc/connect-query";

import { useCompleteLogin } from "#/features/auth/hooks/useCompleteLogin";
import { AuthService } from "#/gen/chat/v1/auth_service_pb";

export const useLoginWithGoogle = (workspaceId: string | null) => {
  const completeLogin = useCompleteLogin(workspaceId);
  return useMutation(AuthService.method.loginWithGoogle, { onSuccess: completeLogin });
};
