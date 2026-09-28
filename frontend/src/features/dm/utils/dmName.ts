import type { DirectMessage } from "#/gen/chat/v1/direct_message_service_pb";

// 参加者（自分以外）の名前を並べる。nameOf には useDisplayName を渡してニックネームを反映する
export const dmName = (
  dm: DirectMessage,
  nameOf: (userId: string, displayName: string) => string,
) => dm.members.map((member) => nameOf(member.userId, member.displayName)).join(", ");
