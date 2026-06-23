/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { Status } from '../domain/types';

export interface StatusConfig {
  label: string;
  color: string;   // Text/accent color hex
  bg: string;      // Background color hex
  border: string;  // Border color hex
  short: string;
}

export const STATUS_CONFIG: Record<Status, StatusConfig> = {
  planned: {
    label: 'Planned',
    color: '#6B7280', // Gray-500
    bg: '#F3F4F6',    // Gray-100
    border: '#D1D5DB',
    short: 'PLN',
  },
  in_progress: {
    label: 'In Progress',
    color: '#3B82F6', // Blue-500
    bg: '#DBEAFE',    // Blue-100
    border: '#93C5FD',
    short: 'DEV',
  },
  review: {
    label: 'In Review',
    color: '#8B5CF6', // Purple-500
    bg: '#EDE9FE',    // Purple-100
    border: '#C4B5FD',
    short: 'REV',
  },
  done: {
    label: 'Done',
    color: '#10B981', // Emerald-500
    bg: '#D1FAE5',    // Emerald-100
    border: '#6EE7B7',
    short: 'DN',
  },
};
