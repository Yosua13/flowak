/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import { NodeType } from '../domain/types';

export interface NodeTypeConfig {
  label: string;
  color: string; // Text / accent hex color
  bg: string;    // Light bg hex color
  border: string; // Border accent color
  darkBg: string; // Dark mode background color
  icon: string;   // Lucide icon name
}

export const NODE_TYPES: Record<NodeType, NodeTypeConfig> = {
  terminator: {
    label: 'Terminator',
    color: '#EF4444', // Red-500
    bg: '#FEE2E2',    // Red-100
    border: '#F87171',
    darkBg: '#7F1D1D',
    icon: 'Play',
  },
  process: {
    label: 'Process',
    color: '#3B82F6', // Blue-500
    bg: '#DBEAFE',    // Blue-100
    border: '#60A5FA',
    darkBg: '#1E3A8A',
    icon: 'Cpu',
  },
  decision: {
    label: 'Decision',
    color: '#F59E0B', // Amber-500
    bg: '#FEF3C7',    // Amber-100
    border: '#FBBF24',
    darkBg: '#78350F',
    icon: 'GitFork',
  },
  actor: {
    label: 'Actor',
    color: '#10B981', // Emerald-500
    bg: '#D1FAE5',    // Emerald-100
    border: '#34D399',
    darkBg: '#064E3B',
    icon: 'User',
  },
  system: {
    label: 'System',
    color: '#8B5CF6', // Purple-500
    bg: '#EDE9FE',    // Purple-100
    border: '#A78BFA',
    darkBg: '#4C1D95',
    icon: 'Database',
  },
};
