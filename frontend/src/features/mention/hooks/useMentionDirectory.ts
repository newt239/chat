import { skipToken, useQuery } from "@connectrpc/connect-query";
import { useParams } from "@tanstack/react-router";
import { useTranslation } from "react-i18next";

import { useChannels } from "#/features/channel/hooks/useChannel";
import { useMembers } from "#/features/member/hooks/useMembers";
import { useUserGroups } from "#/features/userGroup/hooks/useUserGroups";
import { ChannelService } from "#/gen/chat/v1/channel_service_pb";

import { replaceMentionTokens } from "../utils/mentionToken";
import { toPlainText } from "../utils/plainText";

import type { MentionToken } from "../utils/mentionToken";

/** 本文の ID 記法が指すユーザー・グループ・チャンネルを、今の名前で引く */
export const useMentionDirectory = () => {
  const { t } = useTranslation();
  const { workspaceId = null } = useParams({ strict: false });
  const { data: members } = useMembers(workspaceId);
  const { data: groups } = useUserGroups(workspaceId);
  const { data: channels } = useChannels(workspaceId);
  // 未参加の公開チャンネルは一覧にないため、参加できるチャンネルからも探す
  const { data: browsable } = useQuery(
    ChannelService.method.listBrowsableChannels,
    workspaceId === null ? skipToken : { workspaceId },
    { select: (res) => res.channels.flatMap(({ channel }) => channel ?? []) },
  );

  const member = (id: string) => members?.find((item) => item.userId === id);
  const group = (id: string) => groups?.find((item) => item.id === id);
  const channel = (id: string) =>
    channels?.find((item) => item.id === id) ?? browsable?.find((item) => item.id === id);

  // 入力欄や抜粋に出す名前。分からなければ null
  const labelOf = (token: MentionToken) => {
    if (token.kind === "broadcast") {
      return `@${token.id}`;
    }
    if (token.kind === "channel") {
      const found = channel(token.id);
      return found === undefined ? null : `#${found.name}`;
    }
    const found = token.kind === "user" ? member(token.id) : undefined;
    const name = found ? (found.nickname ?? found.displayName) : group(token.id)?.name;
    return name === undefined ? null : `@${name}`;
  };

  // 本文や抜粋に出す名前。分からない宛先は伏せる
  const textOf = (token: MentionToken) =>
    labelOf(token) ??
    (token.kind === "channel"
      ? `#${t("message.mention.unknownChannel")}`
      : `@${t(token.kind === "group" ? "message.mention.unknownGroup" : "message.mention.unknownUser")}`);

  const toText = (body: string) => replaceMentionTokens(body, textOf);

  // 一覧やシートに出す 1 行の抜粋
  const toExcerpt = (body: string) =>
    toPlainText(toText(body)) || t("message.sheet.attachmentOnly");

  return {
    channel,
    group,
    isReady: members !== undefined && groups !== undefined && channels !== undefined,
    labelOf,
    member,
    textOf,
    toExcerpt,
    toText,
    workspaceId,
  };
};
