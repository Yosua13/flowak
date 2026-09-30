import { useEffect, useState } from 'react';
import type { ID, WorkItem } from '../domain/types';
import { apiClient, normalizeApiError } from './apiClient';

export interface StatusHistory { work_item_id: ID; from_status?: WorkItem['status']; to_status: WorkItem['status']; note?: string; created_at: string; }
export interface FacetReview { node_id: ID; node_label: string; role_key: 'uiux' | 'frontend' | 'backend'; readiness: string; status: string; due_date: string; }
export interface ModuleBaseline { module_id: ID; version: number; created_at: string; }
export interface DecisionComment { id: ID; work_item_id?: ID; body: string; resolved_at?: string; created_at: string; }
export interface Evidence { id: ID; kind: 'attachment'; label: string; work_item_id?: ID; node_id?: ID; created_at: string; }
export interface DerivedViewData { work_items: WorkItem[]; history: StatusHistory[]; facet_reviews: FacetReview[]; baselines: ModuleBaseline[]; comments: DecisionComment[]; evidence: Evidence[]; generated_at: string; }

export type CalendarEventKind = 'start' | 'due' | 'review' | 'baseline';
export interface DerivedCalendarEvent { id: string; date: string; title: string; detail: string; kind: CalendarEventKind; }

// Calendar, documents, metrics, and status all consume this same normalized
// response. Keeping the calendar projection here makes that contract testable
// without coupling it to the rendered UI.
export function buildCalendarEvents(data: DerivedViewData): DerivedCalendarEvent[] {
  return [
    ...data.work_items.flatMap((item) => [item.start_date && { id: `${item.id}-start`, date: item.start_date, title: item.title, detail: `${item.key} · mulai`, kind: 'start' as const }, item.due_date && { id: `${item.id}-due`, date: item.due_date, title: item.title, detail: `${item.key} · ${item.status}`, kind: 'due' as const }].filter(Boolean) as DerivedCalendarEvent[]),
    ...data.facet_reviews.map((review) => ({ id: `${review.node_id}-${review.role_key}`, date: review.due_date, title: review.node_label, detail: `${review.role_key.toUpperCase()} · review ${review.readiness}`, kind: 'review' as const })),
    ...data.baselines.map((baseline) => ({ id: `${baseline.module_id}-${baseline.version}`, date: baseline.created_at, title: `Baseline v${baseline.version}`, detail: 'Versi modul tercatat', kind: 'baseline' as const })),
  ];
}

export function derivedViewData(projectId: ID, moduleId?: ID, facetKey?: string, signal?: AbortSignal) {
  const query = new URLSearchParams();
  if (moduleId) query.set('module_id', moduleId);
  if (facetKey) query.set('facet_key', facetKey);
  return apiClient.get<DerivedViewData>(`/projects/${projectId}/derived-view-data${query.size ? `?${query}` : ''}`, signal);
}

export function useDerivedViewData(projectId: ID | null, moduleId?: ID) {
  const [data, setData] = useState<DerivedViewData | null>(null);
  const [loading, setLoading] = useState(Boolean(projectId));
  const [error, setError] = useState<string | null>(null);
  useEffect(() => {
    if (!projectId) { setData(null); setLoading(false); return; }
    const controller = apiClient.abortable();
    setLoading(true); setError(null);
    void derivedViewData(projectId, moduleId, undefined, controller.signal).then(setData).catch((reason) => {
      if (!controller.signal.aborted) setError(normalizeApiError(reason).message);
    }).finally(() => { if (!controller.signal.aborted) setLoading(false); });
    return () => controller.abort();
  }, [projectId, moduleId]);
  return { data, loading, error };
}
