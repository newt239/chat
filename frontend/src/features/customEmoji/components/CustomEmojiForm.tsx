import { useEffect, useState } from "react";

import { Code, ConnectError } from "@connectrpc/connect";
import { IconPhoto } from "@tabler/icons-react";
import { FileTrigger, Form } from "react-aria-components";
import { useTranslation } from "react-i18next";

import { Button } from "#/components/ui/Button/Button";
import { TextField } from "#/components/ui/TextField/TextField";
import { toast } from "#/components/ui/ToastRegion/toast";

import { useCustomEmojiActions } from "../hooks/useCustomEmojiActions";
import { toCustomEmojiValue } from "../utils/customEmoji";
import { EMOJI_IMAGE_TYPES, EmojiImageError } from "../utils/prepareEmojiImage";

const NAME_PATTERN = /^[a-z0-9_-]{1,32}$/u;

// ファイル名から名前の候補を作る。使えない文字は _ にする
const nameFromFileName = (fileName: string) =>
  fileName
    .replace(/\.[^.]+$/u, "")
    .toLowerCase()
    .replaceAll(/[^a-z0-9_-]+/gu, "_")
    .slice(0, 32);

type CustomEmojiFormProps = {
  workspaceId: string;
  // 登録できたときに登録名を受け取る
  onAdded: (name: string) => void;
};

export const CustomEmojiForm = ({ workspaceId, onAdded }: CustomEmojiFormProps) => {
  const { t } = useTranslation();
  const { register } = useCustomEmojiActions(workspaceId);
  const [name, setName] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const isNameValid = NAME_PATTERN.test(name);

  // 差し替えたときとアンマウントしたときに前のプレビューを解放する
  useEffect(
    () => () => {
      if (preview !== null) {
        URL.revokeObjectURL(preview);
      }
    },
    [preview],
  );

  const errorMessage = (() => {
    const { error } = register;
    if (error === null) {
      return null;
    }
    if (error instanceof EmojiImageError) {
      return t(
        error.reason === "type" ? "workspace.emoji.errors.type" : "workspace.emoji.errors.size",
      );
    }
    if (ConnectError.from(error).code === Code.AlreadyExists) {
      return t("workspace.emoji.errors.nameExists");
    }
    return t("workspace.emoji.errors.failed");
  })();

  const selectFile = (selected: File) => {
    setFile(selected);
    setPreview(URL.createObjectURL(selected));
    if (name === "") {
      setName(nameFromFileName(selected.name));
    }
    register.reset();
  };

  return (
    <Form
      className="flex flex-col gap-3"
      onSubmit={(event) => {
        event.preventDefault();
        if (file === null || !isNameValid) {
          return;
        }
        register.mutate(
          { file, name },
          {
            onSuccess: () => {
              toast(t("workspace.emoji.added", { name: toCustomEmojiValue(name) }), {
                tone: "success",
              });
              setName("");
              setFile(null);
              setPreview(null);
              onAdded(name);
            },
          },
        );
      }}
    >
      <TextField
        label={t("workspace.emoji.name")}
        value={name}
        onChange={(value) => {
          setName(value.toLowerCase());
        }}
        description={t("workspace.emoji.nameDescription")}
        errorMessage={name !== "" && !isNameValid ? t("workspace.emoji.errors.name") : undefined}
        className="min-w-0"
      />
      <div className="flex flex-wrap items-center gap-3">
        <FileTrigger
          acceptedFileTypes={EMOJI_IMAGE_TYPES}
          onSelect={(files) => {
            const selected = files?.[0];
            if (selected !== undefined) {
              selectFile(selected);
            }
          }}
        >
          <Button variant="secondary">
            {preview === null ? (
              <IconPhoto aria-hidden className="size-4" />
            ) : (
              <img src={preview} alt="" className="size-5 object-contain" />
            )}
            {file === null ? t("workspace.emoji.selectImage") : t("workspace.emoji.changeImage")}
          </Button>
        </FileTrigger>
        <p role="alert" className="m-0 flex-1 text-caption text-danger">
          {errorMessage}
        </p>
        <Button
          type="submit"
          isDisabled={file === null || !isNameValid}
          isPending={register.isPending}
        >
          {t("workspace.emoji.submit")}
        </Button>
      </div>
    </Form>
  );
};
