import { useTranslation } from "react-i18next";

import { Dialog } from "#/components/ui/Dialog/Dialog";

type MarkdownHelpModalProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
};

const rows = [
  "bold",
  "italic",
  "strikethrough",
  "heading",
  "link",
  "code",
  "quote",
  "list",
  "orderedList",
  "mention",
  "channel",
] as const;

export const MarkdownHelpModal = ({ isOpen, onOpenChange }: MarkdownHelpModalProps) => {
  const { t } = useTranslation();
  const labelOf = (row: (typeof rows)[number]) =>
    row === "mention" || row === "channel"
      ? t(`message.help.${row}`)
      : t(`message.composer.${row}`);

  return (
    <Dialog isOpen={isOpen} onOpenChange={onOpenChange} title={t("message.composer.help")}>
      <p className="m-0 text-muted">{t("message.help.description")}</p>
      <table className="w-full border-collapse text-[13px]">
        <tbody>
          {rows.map((row) => (
            <tr key={row} className="border-b border-border last:border-b-0">
              <td className="py-1.5 pr-3">{labelOf(row)}</td>
              <td className="py-1.5">
                <code className="rounded-sm border border-border bg-sunken px-1 font-mono text-[12.5px]">
                  {t(`message.help.examples.${row}`)}
                </code>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </Dialog>
  );
};
