import { DirectMessageType } from "#/gen/chat/v1/direct_message_service_pb";

import type { DirectMessage } from "#/gen/chat/v1/direct_message_service_pb";

// 1 対 1 は相手の名前、グループは付けた名前か参加者の名前を並べたもの
export const dmName = (dm: DirectMessage) =>
  dm.type === DirectMessageType.DM || dm.name === ""
    ? dm.members.map((member) => member.displayName).join(", ")
    : dm.name;
