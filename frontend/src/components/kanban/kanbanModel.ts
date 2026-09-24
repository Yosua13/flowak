import type { WorkItem, WorkItemStatus } from '../../domain/types';

export const BOARD_STATUSES: WorkItemStatus[] = ['Backlog', 'Ready', 'In Progress', 'In Review', 'Blocked', 'Done'];

const STATUS_TRANSITIONS: Partial<Record<WorkItemStatus, WorkItemStatus[]>> = {
  Backlog: ['Ready'],
  Ready: ['In Progress', 'Blocked'],
  'In Progress': ['Ready', 'In Review', 'Blocked'],
  'In Review': ['In Progress', 'Blocked', 'Done'],
  Blocked: ['Ready', 'In Progress'],
  Done: [],
  Canceled: [],
};

/** Mirrors the server workflow so invalid drops never make a failing request. */
export function canTransitionWorkItem(from: WorkItemStatus, to: WorkItemStatus): boolean {
  return from === to || Boolean(STATUS_TRANSITIONS[from]?.includes(to));
}

export interface KanbanFilters {
  search: string;
  moduleId: string;
  assigneeId: string;
  type: string;
  facet: string;
  priority: string;
  includeCanceled: boolean;
}

export function filterWorkItems(items: WorkItem[], filters: KanbanFilters): WorkItem[] {
  const search = filters.search.trim().toLowerCase();
  return items.filter((item) => {
    if (!filters.includeCanceled && item.status === 'Canceled') return false;
    if (filters.moduleId !== 'all' && item.module_id !== filters.moduleId) return false;
    if (filters.assigneeId !== 'all' && item.assignee_id !== filters.assigneeId) return false;
    if (filters.type !== 'all' && item.type !== filters.type) return false;
    if (filters.facet !== 'all' && item.facet_key !== filters.facet) return false;
    if (filters.priority !== 'all' && item.priority !== filters.priority) return false;
    return !search || `${item.key} ${item.title} ${item.description || ''}`.toLowerCase().includes(search);
  });
}

export function columnMetrics(items: WorkItem[], status: WorkItemStatus, today = new Date()) {
  const cards = items.filter((item) => item.status === status);
  const start = new Date(today.getFullYear(), today.getMonth(), today.getDate()).getTime();
  return {
    count: cards.length,
    points: cards.reduce((total, item) => total + (item.points || 0), 0),
    overdue: cards.filter((item) => item.due_date && new Date(`${item.due_date.slice(0, 10)}T00:00:00`).getTime() < start && item.status !== 'Done' && item.status !== 'Canceled').length,
  };
}

export function itemByKey(items: WorkItem[], key: string | null): WorkItem | undefined {
  return key ? items.find((item) => item.key === key) : undefined;
}
