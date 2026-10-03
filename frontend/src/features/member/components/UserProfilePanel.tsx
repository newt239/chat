import { formatTime } from "@chat/i18n/format";
import { IconMessage, IconTag } from "@tabler/icons-react";
import { useNavigate } from "@tanstack/react-router";
import { useAtomValue } from "jotai";
import { Link } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Avatar } from "#/components/ui/Avatar/Avatar";
import { Button } from "#/components/ui/Button/Button";
import { Skeleton } from "#/components/ui/Skeleton/Skeleton";
import { focusRing } from "#/components/ui/styles/styles";
import { useCreateDM } from "#/features/dm/hooks/useDM";
import { useMembers } from "#/features/member/hooks/useMembers";
import { workspaceRoleKeys } from "#/features/member/utils/workspaceRoleKeys";
import { usePreferences } from "#/hooks/usePreferences";
import { myUserIdAtom } from "#/providers/store/auth";

import { useUserNote } from "../hooks/useUserNote";
import { displayUrl, linkIconOf } from "../utils/linkIcon";
import { UserNoteEditor } from "./UserNoteEditor";

type UserProfilePanelProps = {
  workspaceId: string;
  userId: string;
};

export const UserProfilePanel = ({ workspaceId, userId }: UserProfilePanelProps) => {
  const { t } = useTranslation();
  const { data: members, isLoading, isError } = useMembers(workspaceId);
  const myId = useAtomValue(myUserIdAtom);
  const { locale } = usePreferences();
  const createDM = useCreateDM();
  const navigate = useNavigate();
  const member = members?.find((candidate) => candidate.userId === userId);
  const isMe = myId === userId;
  const { data: note, isLoading: isLoadingNote } = useUserNote(isMe ? null : userId);

  const startDM = async () => {
    const { directMessage } = await createDM.mutateAsync({ userId, workspaceId });
    if (directMessage !== undefined) {
      void navigate({
        params: { channelId: directMessage.id, workspaceId },
        to: "/app/$workspaceId/$channelId",
      });
    }
  };

  if (isLoading) {
    return (
      <div className="flex flex-col gap-3 p-4">
        <Skeleton className="size-18 rounded-xl" />
        <Skeleton className="h-5 w-40" />
      </div>
    );
  }

  if (isError || member === undefined) {
    return (
      <p className="m-0 p-4 text-caption text-muted">
        {isError ? t("member.profile.loadFailed") : t("member.profile.notFound")}
      </p>
    );
  }

  return (
    <div className="flex min-h-full flex-col bg-surface font-sans text-text">
      <section className="flex flex-col gap-2.5 border-b border-border px-4 pt-4 pb-3.5">
        <Avatar name={member.displayName} src={member.avatarUrl} size={72} />
        <div className="flex flex-col gap-0.5">
          <h3 className="m-0 flex items-center gap-1.5 text-heading font-bold">
            {member.nickname ?? member.displayName}
            {member.nickname !== undefined && (
              <IconTag
                aria-label={t("member.note.nickname")}
                role="img"
                className="size-4 text-muted"
              />
            )}
          </h3>
          {member.nickname !== undefined && (
            <span className="text-label font-normal text-muted">
              {t("member.note.realName", { name: member.displayName })}
            </span>
          )}
        </div>
        <dl className="m-0 grid grid-cols-[auto_1fr] gap-x-3.5 gap-y-1 text-label font-normal">
          <dt className="text-muted">{t("member.profile.role")}</dt>
          <dd className="m-0">{t(workspaceRoleKeys[member.role])}</dd>
          <dt className="text-muted">{t("member.profile.email")}</dt>
          <dd className="m-0 truncate">{member.email}</dd>
          {member.timezone !== "" && (
            <>
              <dt className="text-muted">{t("member.profile.localTime")}</dt>
              <dd className="m-0 truncate">
                {formatTime(new Date(), locale, member.timezone)}
                <span className="ml-1.5 text-muted">{member.timezone}</span>
              </dd>
            </>
          )}
        </dl>
        {!isMe && (
          <div className="flex gap-1.5">
            <Button
              isPending={createDM.isPending}
              onPress={() => {
                void startDM();
              }}
            >
              <IconMessage aria-hidden />
              {t("member.profile.message")}
            </Button>
          </div>
        )}
      </section>
      {member.links.length > 0 && (
        <section className="flex flex-col gap-1.5 border-b border-border px-4 py-3">
          <h4 className="m-0 text-xs font-semibold text-muted">{t("member.profile.links")}</h4>
          <ul className="m-0 flex list-none flex-col gap-1 p-0">
            {member.links.map((url) => {
              const SiteIcon = linkIconOf(url);
              return (
                <li key={url} className="flex min-w-0 items-center gap-1.5 text-body-sm">
                  <SiteIcon aria-hidden className="size-4 shrink-0 text-muted" />
                  <Link
                    href={url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className={`truncate rounded-sm text-accent-text no-underline data-hovered:underline ${focusRing}`}
                  >
                    {displayUrl(url)}
                  </Link>
                </li>
              );
            })}
          </ul>
        </section>
      )}
      {!isMe && !isLoadingNote && (
        <UserNoteEditor
          key={note?.updatedAt?.seconds.toString() ?? "new"}
          targetUserId={userId}
          displayName={member.displayName}
          initialNickname={note?.nickname ?? ""}
          initialMemo={note?.memo ?? ""}
        />
      )}
      {member.bio !== undefined && member.bio.length > 0 && (
        <section className="flex flex-col gap-2 px-4 py-3">
          <h4 className="m-0 text-xs font-semibold text-muted">{t("member.profile.bio")}</h4>
          <p className="m-0 text-body-sm whitespace-pre-wrap">{member.bio}</p>
        </section>
      )}
    </div>
  );
};
