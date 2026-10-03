import { Keyboard } from "react-aria-components";
import { useTranslation } from "react-i18next";

const shortcuts = [
  ["search", ["⌘ / Ctrl", "K"]],
  ["settings", ["⌘ / Ctrl", ","]],
  ["nextUnread", ["⌥ / Alt", "Shift", "↓"]],
  ["send", ["Enter"]],
  ["newline", ["Shift", "Enter"]],
  ["close", ["Esc"]],
  ["newTab", ["⌘ / Ctrl", "Click"]],
] as const;

export const ShortcutSettings = () => {
  const { t } = useTranslation();
  return (
    <dl className="m-0 grid grid-cols-[1fr_auto] items-center gap-x-4 gap-y-2.5">
      {shortcuts.map(([name, keys]) => (
        <div key={name} className="contents">
          <dt className="text-body">{t(`settings.shortcuts.${name}`)}</dt>
          <dd className="m-0 flex gap-1">
            {keys.map((key) => (
              <Keyboard
                key={key}
                className="rounded-sm border border-border-strong bg-sunken px-1.5 font-mono text-xs leading-5"
              >
                {key}
              </Keyboard>
            ))}
          </dd>
        </div>
      ))}
    </dl>
  );
};
