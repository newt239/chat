import { useChannelLinks } from "../hooks/useChannelLinks";
import { ChannelLinkDialog } from "./ChannelLinkDialog";

type ChannelLinkDialogLoaderProps = {
  channelId: string;
  // null なら追加
  linkId: string | null;
  onClose: () => void;
};

// URL の ID から編集するリンクを引いてダイアログを開く。入力欄の初期値にするため、読み込めてから描く
export const ChannelLinkDialogLoader = ({
  channelId,
  linkId,
  onClose,
}: ChannelLinkDialogLoaderProps) => {
  const { data } = useChannelLinks(channelId);
  const link = data?.links.find((candidate) => candidate.id === linkId) ?? null;

  if (!data?.canEdit || (linkId !== null && link === null)) {
    return null;
  }
  return <ChannelLinkDialog key={linkId} channelId={channelId} link={link} onClose={onClose} />;
};
