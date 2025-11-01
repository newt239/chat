import { IconHash, IconLock } from "@tabler/icons-react";

type ChannelNameProps = {
  name: string;
  isPrivate: boolean;
  isBold?: boolean;
};

export const ChannelName = ({ name, isPrivate, isBold = false }: ChannelNameProps) => {
  return (
    <div>
      <div className="flex items-center gap-2">
        {isPrivate ? (
          <IconLock size={20} title="プライベートチャンネル" />
        ) : (
          <IconHash size={20} title="パブリックチャンネル" />
        )}
        <span className={isBold ? "font-bold" : ""}>{name}</span>
      </div>
    </div>
  );
};
