import type { Module, WorkItem } from '../domain/types';
import type { StatusHistory } from './derivedViewData';

export interface DerivedMetrics { throughput: number; cycleTimeHours: number | null; blockedTimeHours: number; agingWipHours: number | null; overdue: number; reviewRework: number; completeness: number; traceability: number; }

const done = (status: WorkItem['status']) => status === 'Done' || status === 'Canceled';
const milliseconds = (value: string) => new Date(value).getTime();

export function calculateDerivedMetrics(items: WorkItem[], history: StatusHistory[], module: Module, now = new Date()): DerivedMetrics {
  const byItem = new Map<string, StatusHistory[]>();
  history.forEach((entry) => byItem.set(entry.work_item_id, [...(byItem.get(entry.work_item_id) || []), entry].sort((a, b) => milliseconds(a.created_at) - milliseconds(b.created_at))));
  const cycleHours: number[] = [];
  let throughput = 0, blockedMilliseconds = 0, reviewRework = 0;
  for (const item of items) {
    const events = byItem.get(item.id) || [];
    const created = events[0]?.created_at || item.created_at;
    let blockedAt: string | null = null;
    events.forEach((event) => {
      if (event.to_status === 'Done') { throughput++; if (created) cycleHours.push((milliseconds(event.created_at) - milliseconds(created)) / 3_600_000); }
      if (event.to_status === 'Blocked') blockedAt = event.created_at;
      if (blockedAt && event.from_status === 'Blocked' && event.to_status !== 'Blocked') { blockedMilliseconds += milliseconds(event.created_at) - milliseconds(blockedAt); blockedAt = null; }
      if (event.from_status === 'In Review' && event.to_status === 'In Progress') reviewRework++;
    });
    if (blockedAt) blockedMilliseconds += now.getTime() - milliseconds(blockedAt);
  }
  const wip = items.filter((item) => !done(item.status));
  const aging = wip.map((item) => (now.getTime() - milliseconds(item.created_at || now.toISOString())) / 3_600_000);
  const today = now.toISOString().slice(0, 10);
  const readiness = module.nodes.flatMap((node) => ['uiux', 'frontend', 'backend'].map((facet) => node.roles[facet as keyof typeof node.roles])).filter(Boolean);
  return {
    throughput,
    cycleTimeHours: cycleHours.length ? cycleHours.reduce((sum, value) => sum + value, 0) / cycleHours.length : null,
    blockedTimeHours: blockedMilliseconds / 3_600_000,
    agingWipHours: aging.length ? aging.reduce((sum, value) => sum + value, 0) / aging.length : null,
    overdue: items.filter((item) => Boolean(item.due_date && item.due_date.slice(0, 10) < today && !done(item.status))).length,
    reviewRework,
    completeness: readiness.length ? readiness.filter((facet) => facet?.readiness === 'done' || facet?.status === 'done').length / readiness.length : 0,
    traceability: items.length ? items.filter((item) => item.node_id).length / items.length : 0,
  };
}
