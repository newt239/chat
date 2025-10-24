import { useState, useCallback } from 'react';

export type MessageInputMode = 'edit' | 'preview';

export function useMessageInputMode(initialMode: MessageInputMode = 'edit') {
  const [mode, setMode] = useState<MessageInputMode>(initialMode);

  const toggleMode = useCallback(() => {
    setMode((prev) => (prev === 'edit' ? 'preview' : 'edit'));
  }, []);

  const setEditMode = useCallback(() => {
    setMode('edit');
  }, []);

  const setPreviewMode = useCallback(() => {
    setMode('preview');
  }, []);

  return {
    mode,
    isEditMode: mode === 'edit',
    isPreviewMode: mode === 'preview',
    toggleMode,
    setEditMode,
    setPreviewMode,
  };
}
