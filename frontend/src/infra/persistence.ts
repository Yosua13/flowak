/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

/** Local storage is deliberately limited to user preferences; server state never lives here. */
export interface Preferences { darkMode: boolean; }
const PREFERENCES_KEY = 'flowak_preferences';

export const persistenceAdapter = {
  loadPreferences(): Preferences {
    try { return { darkMode: JSON.parse(localStorage.getItem(PREFERENCES_KEY) ?? '{}').darkMode ?? true }; }
    catch { return { darkMode: true }; }
  },
  savePreferences(preferences: Preferences): void {
    localStorage.setItem(PREFERENCES_KEY, JSON.stringify(preferences));
  },
};
