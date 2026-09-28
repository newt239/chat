import { findThemePreset, themePresetNames, themePresets } from "@chat/design-tokens";
import { useAtom } from "jotai";
import { Button } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { SegmentedControl } from "#/components/ui/SegmentedControl";
import { Slider } from "#/components/ui/Slider";
import { focusRing } from "#/components/ui/styles";
import { preferencesAtom } from "#/providers/store/preferences";

import { useUpdatePreferences } from "../hooks/usePreferences";
import { SettingRow } from "./SettingRow";
import { ThemePreview } from "./ThemePreview";

// 色相の帯。彩度と明度を固定して 30° ごとに並べる
const hueTrack = `linear-gradient(90deg, ${Array.from(
  { length: 13 },
  (_, index) => `oklch(0.65 0.14 ${index * 30})`,
).join(", ")})`;

export const ThemeSettings = () => {
  const { t } = useTranslation();
  const [preferences, setPreferences] = useAtom(preferencesAtom);
  const { mode, theme } = preferences;
  const updatePreferences = useUpdatePreferences();
  const currentPreset = findThemePreset(theme);

  return (
    <div className="flex flex-col gap-5">
      <section className="flex flex-col gap-2.5">
        <h3 className="m-0 text-body-strong">{t("preferences.theme.presetsTitle")}</h3>
        <div className="grid grid-cols-[repeat(auto-fill,minmax(140px,1fr))] gap-2.5">
          {themePresetNames.map((name) => (
            <Button
              key={name}
              aria-pressed={currentPreset === name}
              onPress={() => {
                updatePreferences({ theme: themePresets[name] });
              }}
              className={`flex cursor-pointer flex-col overflow-hidden rounded-[10px] border border-border bg-surface text-left aria-pressed:border-accent aria-pressed:ring-1 aria-pressed:ring-accent ${focusRing}`}
            >
              <ThemePreview theme={themePresets[name]} />
              <span className="flex items-center justify-between px-2.5 py-1.5 text-[13px] font-semibold">
                {t(`preferences.theme.presets.${name}`)}
                <small className="font-mono text-[11px] font-normal text-muted">
                  {themePresets[name].hue}°
                </small>
              </span>
            </Button>
          ))}
        </div>
      </section>

      <section className="flex flex-col">
        <h3 className="m-0 mb-1 text-body-strong">
          {t("preferences.theme.custom")}
          {currentPreset === undefined && (
            <span className="ml-2 text-caption font-normal text-accent-text">
              {t("preferences.theme.customActive")}
            </span>
          )}
        </h3>
        <div className="flex flex-col gap-3 py-2">
          <Slider
            label={t("preferences.theme.hue")}
            value={theme.hue}
            minValue={0}
            maxValue={359}
            step={1}
            trackBackground={hueTrack}
            onChange={(hue) => {
              setPreferences({ ...preferences, theme: { ...theme, hue } });
            }}
            onChangeEnd={(hue) => {
              updatePreferences({ theme: { ...theme, hue } });
            }}
          />
          <Slider
            label={t("preferences.theme.chroma")}
            value={theme.chroma}
            minValue={0}
            maxValue={0.3}
            step={0.005}
            onChange={(chroma) => {
              setPreferences({ ...preferences, theme: { ...theme, chroma } });
            }}
            onChangeEnd={(chroma) => {
              updatePreferences({ theme: { ...theme, chroma } });
            }}
          />
        </div>
        <SettingRow title={t("preferences.theme.sidebar.title")} description={null}>
          <SegmentedControl
            label={t("preferences.theme.sidebar.title")}
            options={[
              { label: t("preferences.theme.sidebar.tinted"), value: "tinted" },
              { label: t("preferences.theme.sidebar.light"), value: "light" },
            ]}
            value={theme.sidebar}
            onChange={(sidebar) => {
              updatePreferences({ theme: { ...theme, sidebar } });
            }}
          />
        </SettingRow>
        <SettingRow title={t("preferences.mode.title")} description={null}>
          <SegmentedControl
            label={t("preferences.mode.title")}
            options={(["system", "light", "dark"] as const).map((value) => ({
              label: t(`preferences.mode.${value}`),
              value,
            }))}
            value={mode}
            onChange={(value) => {
              updatePreferences({ mode: value });
            }}
          />
        </SettingRow>
      </section>

      <section className="flex flex-col gap-2">
        <h3 className="m-0 text-body-strong">{t("preferences.theme.preview")}</h3>
        <div className="overflow-hidden rounded-[10px] border border-border [&>span]:h-28">
          <ThemePreview theme={theme} />
        </div>
      </section>
    </div>
  );
};
