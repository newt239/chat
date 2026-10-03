import { useContext } from "react";

import { IconChevronLeft } from "@tabler/icons-react";
import { useTranslation } from "react-i18next";

import { IconButton } from "#/components/ui/IconButton/IconButton";

import { MobileBackContext } from "./mobileBackContext";

// モバイルで積み重ねた画面の見出しに出す「戻る」。それ以外では何も出さない
export const BackButton = () => {
  const { t } = useTranslation();
  const back = useContext(MobileBackContext);
  if (back === null) {
    return null;
  }
  return (
    <IconButton label={t("shell.back")} onPress={back} className="-ml-2 size-10 [&_svg]:size-5.25">
      <IconChevronLeft />
    </IconButton>
  );
};
