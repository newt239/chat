const {
  VITE_FIREBASE_API_KEY: apiKey,
  VITE_FIREBASE_APP_ID: appId,
  VITE_FIREBASE_MESSAGING_SENDER_ID: messagingSenderId,
  VITE_FIREBASE_PROJECT_ID: projectId,
  VITE_FIREBASE_VAPID_KEY: vapidKey,
} = import.meta.env;

// VITE_FIREBASE_* が揃っていなければプッシュ通知を使わない（登録の UI も出さない）
export const firebaseConfig =
  apiKey && appId && messagingSenderId && projectId && vapidKey
    ? { options: { apiKey, appId, messagingSenderId, projectId }, vapidKey }
    : null;
