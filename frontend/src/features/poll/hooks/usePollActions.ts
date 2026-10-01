import { useMutation } from "@connectrpc/connect-query";

import { PollService } from "#/gen/chat/v1/poll_service_pb";

/** 投票と締め切り。集計の変化はメッセージの更新として WebSocket で届く */
export const usePollActions = () => ({
  close: useMutation(PollService.method.closePoll),
  vote: useMutation(PollService.method.vote),
});
