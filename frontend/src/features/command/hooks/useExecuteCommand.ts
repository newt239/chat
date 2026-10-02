import { useMutation } from "@connectrpc/connect-query";

import { CommandService } from "#/gen/chat/v1/command_service_pb";

/** コマンドを実行する。応答は公式アプリの投稿として WebSocket で届く */
export const useExecuteCommand = () => useMutation(CommandService.method.executeCommand);
