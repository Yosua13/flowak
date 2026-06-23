/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { Module } from '../domain/types';

export interface PersistenceAdapter {
  load(): Promise<Module[] | null>;
  save(modules: Module[]): Promise<void>;
}

const DB_NAME = 'flowak_db';
const DB_VERSION = 1;
const STORE_NAME = 'modules_store';

function openDatabase(): Promise<IDBDatabase> {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_NAME, DB_VERSION);

    request.onupgradeneeded = (event) => {
      const db = request.result;
      if (!db.objectStoreNames.contains(STORE_NAME)) {
        db.createObjectStore(STORE_NAME, { keyPath: 'id' });
      }
    };

    request.onsuccess = () => {
      resolve(request.result);
    };

    request.onerror = () => {
      reject(request.error);
    };
  });
}

export const persistenceAdapter: PersistenceAdapter = {
  async load(): Promise<Module[] | null> {
    try {
      const db = await openDatabase();
      return new Promise((resolve, reject) => {
        const transaction = db.transaction(STORE_NAME, 'readonly');
        const store = transaction.objectStore(STORE_NAME);
        const request = store.getAll();

        request.onsuccess = () => {
          const result = request.result as Module[];
          resolve(result.length > 0 ? result : null);
        };

        request.onerror = () => {
          resolve(null);
        };
      });
    } catch (e) {
      console.warn('IndexedDB load failed, falling back to memory:', e);
      return null;
    }
  },

  async save(modules: Module[]): Promise<void> {
    try {
      const db = await openDatabase();
      return new Promise((resolve, reject) => {
        const transaction = db.transaction(STORE_NAME, 'readwrite');
        const store = transaction.objectStore(STORE_NAME);

        // Clear existing store to sync completely
        const clearRequest = store.clear();

        clearRequest.onsuccess = () => {
          let count = 0;
          if (modules.length === 0) {
            resolve();
            return;
          }

          modules.forEach((mod) => {
            const addReq = store.put(mod);
            addReq.onsuccess = () => {
              count++;
              if (count === modules.length) {
                resolve();
              }
            };
            addReq.onerror = () => {
              reject(addReq.error);
            };
          });
        };

        clearRequest.onerror = () => {
          reject(clearRequest.error);
        };
      });
    } catch (e) {
      console.warn('IndexedDB save failed:', e);
    }
  },
};
