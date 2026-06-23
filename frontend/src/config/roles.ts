/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { RoleKey } from '../domain/types';

export interface RoleConfig {
  label: string;
  color: string;   // Accent hex
  bg: string;      // Badge bg hex
  border: string;  // Border hex
  short: string;   // Shorthand name
}

export const ROLES: Record<RoleKey, RoleConfig> = {
  uiux: {
    label: 'UI/UX Designer',
    color: '#EC4899', // Pink-500
    bg: '#FCE7F3',    // Pink-100
    border: '#F472B6',
    short: 'UX',
  },
  frontend: {
    label: 'Frontend Developer',
    color: '#06B6D4', // Cyan-500
    bg: '#CFFAFE',    // Cyan-100
    border: '#22D3EE',
    short: 'FE',
  },
  backend: {
    label: 'Backend Developer',
    color: '#F97316', // Orange-500
    bg: '#FFEDD5',    // Orange-100
    border: '#FB923C',
    short: 'BE',
  },
};
