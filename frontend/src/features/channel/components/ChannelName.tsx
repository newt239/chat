import { IconHash, IconLock } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

type ChannelNameProps = {
  name: string;
  isPrivate: boolean;
};

// 名前は階層のパス（frontend/web）。親の階層は控えめに表示し、末尾を強調する
export const ChannelName = ({ name, isPrivate }: ChannelNameProps) => {
  const { t } = useTranslation();
  const separator = name.lastIndexOf("/");
  const Icon = isPrivate ? IconLock : IconHash;
  return (
    <>
      <Icon
        aria-label={isPrivate ? t("shell.channel.private") : t("shell.channel.public")}
        role="img"
      />
      <span className="min-w-0 flex-1 truncate">
        {separator !== -1 && (
          <span className="font-normal opacity-70">
            {name.slice(0, separator).replaceAll("/", " / ")} /{" "}
          </span>
        )}
        {name.slice(separator + 1)}
      </span>
    </>
  );
};
