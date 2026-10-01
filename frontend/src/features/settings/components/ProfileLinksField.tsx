import { IconPlus, IconX } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { IconButton } from "#/components/ui/IconButton/IconButton";
import { TextField } from "#/components/ui/TextField/TextField";

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
      <small className="text-caption text-muted">
        {t("settings.profile.linksDescription", { max: MAX_PROFILE_LINKS })}
      </small>
      {value.map((url, index) => (
        // 並び替えないため、位置を key にする
        // oxlint-disable-next-line react/no-array-index-key
        <div key={index} className="flex items-end gap-1.5">
          <TextField
            className="min-w-0 flex-1"
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
      ))}
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
