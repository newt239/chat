import { IconPlus, IconX } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { TextField } from "#/components/ui/TextField/TextField";
import { linkIconOf } from "#/features/member/utils/linkIcon";

import { MAX_PROFILE_LINKS } from "../utils/profileLinks";

type ProfileLinksFieldProps = {
  value: string[];
  onChange: (value: string[]) => void;
};

/** プロフィールのリンクの URL を並べて編集する */
export const ProfileLinksField = ({ value, onChange }: ProfileLinksFieldProps) => {
  const { t } = useTranslation();

  return (
    <fieldset className="m-0 flex flex-col gap-2 border-0 p-0">
      <legend className="mb-1 text-xs font-semibold text-muted">
        {t("settings.profile.links")}
      </legend>
      {value.map((url, index) => {
        const SiteIcon = linkIconOf(url);
        return (
          // oxlint-disable-next-line react/no-array-index-key -- 並び替えないため位置を key にする
          <div key={index} className="flex items-center gap-1.5">
            <SiteIcon aria-hidden className="size-4 shrink-0 text-muted" />
            {/* ラベルは読み上げにだけ使い、画面には出さない */}
            <TextField
              className="min-w-0 flex-1 [&>label]:sr-only"
              label={t("settings.profile.linkUrl", { number: index + 1 })}
              type="url"
              placeholder="https://"
              value={url}
              onChange={(next) => {
                onChange(value.map((current, i) => (i === index ? next : current)));
              }}
            />
            <IconButton
              label={t("settings.profile.removeLink", { number: index + 1 })}
              onPress={() => {
                onChange(value.filter((_, i) => i !== index));
              }}
            >
              <IconX />
            </IconButton>
          </div>
        );
      })}
      {value.length < MAX_PROFILE_LINKS && (
        <Button
          variant="ghost"
          size="sm"
          className="self-start"
          onPress={() => {
            onChange([...value, ""]);
          }}
        >
          <IconPlus aria-hidden />
          {t("settings.profile.addLink")}
        </Button>
      )}
    </fieldset>
  );
};
