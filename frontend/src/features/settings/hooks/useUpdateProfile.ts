import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";

import { UserService } from "#/gen/chat/v1/user_service_pb";

export const useUpdateProfile = () => {
  const queryClient = useQueryClient();

  return useMutation(UserService.method.updateMe, {
    onSuccess: () =>
      queryClient.invalidateQueries({
        queryKey: createConnectQueryKey({
          cardinality: "finite",
          schema: UserService.method.getMe,
        }),
      }),
  });
};
