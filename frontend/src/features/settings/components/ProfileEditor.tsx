import { Skeleton } from "#/components/ui/Skeleton/Skeleton";

import { useMe } from "../hooks/useMe";
import { ProfileForm } from "./ProfileForm";
import { TimezoneSettings } from "./TimezoneSettings";

// 自分のプロフィールパネルに出す編集フォーム。タイムゾーンは保存ボタンを待たずに反映する
export const ProfileEditor = () => {
  const { data: me } = useMe();
  if (!me) {
    return <Skeleton className="m-4 h-16 w-48" />;
  }
  return (
    <>
      <ProfileForm key={me.displayName} me={me} />
      <section className="border-t border-border px-4">
        <TimezoneSettings />
      </section>
    </>
  );
};
