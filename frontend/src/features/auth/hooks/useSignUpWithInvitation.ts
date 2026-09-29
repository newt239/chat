import { useMutation } from "@connectrpc/connect-query";

import { useCompleteLogin } from "#/features/auth/hooks/useCompleteLogin";
import { AuthService } from "#/gen/chat/v1/auth_service_pb";

export const useSignUpWithInvitation = () => {
  const completeLogin = useCompleteLogin();
  return useMutation(AuthService.method.signUpWithInvitation, { onSuccess: completeLogin });
};
