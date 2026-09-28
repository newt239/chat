import { createConnectQueryKey, useMutation } from "@connectrpc/connect-query";
import { useQueryClient } from "@tanstack/react-query";
import { useAtom } from "jotai";

import { UserService } from "#/gen/chat/v1/user_service_pb";
import { authAtom } from "#/providers/store/auth";

export const useUpdateProfile = () => {
  const queryClient = useQueryClient();
  const [auth, setAuth] = useAtom(authAtom);

  return useMutation(UserService.method.updateMe, {
    onSuccess: async ({ user }) => {
      setAuth({ ...auth, user: user ?? auth.user });
      await queryClient.invalidateQueries({
        queryKey: createConnectQueryKey({
          cardinality: "finite",
          schema: UserService.method.getMe,
        }),
      });
    },
  });
};
