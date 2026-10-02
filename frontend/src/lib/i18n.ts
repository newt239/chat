import { createI18n } from "@chat/i18n";
import { initReactI18next } from "react-i18next";

import { storedPreferencesAtom } from "#/providers/store/preferences";
import { store } from "#/providers/store/store";

export const i18n = createI18n(store.get(storedPreferencesAtom).locale, [initReactI18next]);
