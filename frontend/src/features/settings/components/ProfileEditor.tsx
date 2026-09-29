import { Skeleton } from "#/components/ui/Skeleton/Skeleton";

import { useMe } from "../hooks/useMe";
import { ProfileForm } from "./ProfileForm";

// 自分のプロフィールパネルに出す編集フォーム
export const ProfileEditor = () => {
  const { data: me } = useMe();
  if (!me) {
    return <Skeleton className="m-4 h-16 w-48" />;
  }
  return <ProfileForm key={me.displayName} me={me} />;
};
