/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { create } from 'zustand';

export type DialogTone = 'danger' | 'warning' | 'success' | 'info';

export interface PromptDialogOptions {
  title: string;
  subtitle?: string;
  placeholder?: string;
  initialValue?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  badge?: string;
  tone?: DialogTone;
  required?: boolean;
  multiline?: boolean;
  quickOptions?: string[];
  icon?: 'blocked' | 'check' | 'warning' | 'edit' | 'info';
}

export interface ConfirmDialogOptions {
  title: string;
  message: string;
  subtitle?: string;
  confirmLabel?: string;
  cancelLabel?: string;
  badge?: string;
  tone?: DialogTone;
  icon?: 'trash' | 'archive' | 'warning' | 'info';
}

export type ActiveDialogConfig =
  | ({ type: 'prompt'; resolve: (val: string | null) => void } & PromptDialogOptions)
  | ({ type: 'confirm'; resolve: (val: boolean) => void } & ConfirmDialogOptions);

interface DialogState {
  activeDialog: ActiveDialogConfig | null;
  openPrompt: (options: PromptDialogOptions) => Promise<string | null>;
  openConfirm: (options: ConfirmDialogOptions) => Promise<boolean>;
  close: () => void;
}

export const useDialogStore = create<DialogState>((set, get) => ({
  activeDialog: null,
  openPrompt: (options) => {
    return new Promise<string | null>((resolve) => {
      // If a previous dialog was pending, cancel it safely
      const prev = get().activeDialog;
      if (prev) {
        if (prev.type === 'prompt') prev.resolve(null);
        else prev.resolve(false);
      }

      set({
        activeDialog: {
          ...options,
          type: 'prompt',
          resolve: (val) => {
            set({ activeDialog: null });
            resolve(val);
          },
        },
      });
    });
  },
  openConfirm: (options) => {
    return new Promise<boolean>((resolve) => {
      const prev = get().activeDialog;
      if (prev) {
        if (prev.type === 'prompt') prev.resolve(null);
        else prev.resolve(false);
      }

      set({
        activeDialog: {
          ...options,
          type: 'confirm',
          resolve: (val) => {
            set({ activeDialog: null });
            resolve(val);
          },
        },
      });
    });
  },
  close: () => {
    const current = get().activeDialog;
    if (current) {
      if (current.type === 'prompt') current.resolve(null);
      else current.resolve(false);
    }
    set({ activeDialog: null });
  },
}));

/**
 * Imperative helper to trigger modern popups from anywhere
 */
export const dialog = {
  prompt: (options: PromptDialogOptions): Promise<string | null> => {
    return useDialogStore.getState().openPrompt(options);
  },
  confirm: (options: ConfirmDialogOptions): Promise<boolean> => {
    return useDialogStore.getState().openConfirm(options);
  },
  close: () => {
    useDialogStore.getState().close();
  },
};
