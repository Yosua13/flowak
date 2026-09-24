import type { ID, WorkItem, WorkItemArtifacts, WorkItemStatus, WorkItemType } from '../domain/types';
import { apiClient } from './apiClient';

export interface WorkItemInput {
  module_id?: ID; node_id?: ID; facet_key?: string; parent_id?: ID;
  type: WorkItemType; title: string; description?: string;
  priority?: WorkItem['priority']; points?: WorkItem['points']; assignee_id?: ID;
  status?: WorkItemStatus; start_date?: string; due_date?: string; blocked_reason?: string;
}

export interface WorkItemPage { items: WorkItem[]; next_cursor: string; sort: 'newest' | 'oldest'; }
export type ArtifactKind = 'checklist' | 'watchers' | 'links' | 'attachments';

/** API contract only; state remains server-authoritative. */
export const workItemsApi = {
  list(projectId: ID, query = ''): Promise<WorkItem[]> {
    return apiClient.get(`/projects/${projectId}/work-items${query ? `?${query}` : ''}`);
  },
  page(projectId: ID, cursor = '', sort: WorkItemPage['sort'] = 'newest'): Promise<WorkItemPage> {
    const params = new URLSearchParams({ page: '1', limit: '100', sort });
    if (cursor) params.set('cursor', cursor);
    return apiClient.get(`/projects/${projectId}/work-items?${params}`);
  },
  async listAll(projectId: ID): Promise<WorkItem[]> {
    const items: WorkItem[] = [];
    const seenCursors = new Set<string>();
    const seenItems = new Set<ID>();
    let cursor = '';
    do {
      const page = await workItemsApi.page(projectId, cursor);
      for (const item of page.items) {
        if (!seenItems.has(item.id)) { items.push(item); seenItems.add(item.id); }
      }
      cursor = page.next_cursor;
      if (cursor && seenCursors.has(cursor)) throw new Error('Cursor work item berulang. Muat ulang board.');
      if (cursor) seenCursors.add(cursor);
    } while (cursor);
    return items;
  },
  create(projectId: ID, input: WorkItemInput): Promise<WorkItem> {
    return apiClient.post(`/projects/${projectId}/work-items`, input);
  },
  transition(key: string, status: WorkItemStatus, row_version: number, note = ''): Promise<WorkItem> {
    const resolution = status === 'Done' ? note : undefined;
    return apiClient.post(`/work-items/${key}/transitions`, { status, row_version, note, resolution });
  },
  update(key: string, input: Record<string, unknown> & { row_version: number }): Promise<WorkItem> {
    return apiClient.request(`/work-items/${key}`, { method: 'PATCH', body: input });
  },
  comments(key: string) {
    return apiClient.get<Array<{ id: ID; parent_id?: ID; author_id: ID; body: string; resolved_at?: string; created_at: string; updated_at: string }>>(`/work-items/${key}/comments`);
  },
  comment(key: string, body: string) {
    return apiClient.post<{ id: ID; body: string }>(`/work-items/${key}/comments`, { body });
  },
  artifacts(key: string): Promise<WorkItemArtifacts> {
    return apiClient.get(`/work-items/${encodeURIComponent(key)}/artifacts`);
  },
  createArtifact(key: string, kind: ArtifactKind, input: Record<string, unknown>) {
    return apiClient.post<{ id: ID }>(`/work-items/${encodeURIComponent(key)}/artifacts/${kind}`, input);
  },
  updateArtifact(key: string, kind: ArtifactKind, id: ID, input: Record<string, unknown>) {
    return apiClient.request<{ id: ID }>(`/work-items/${encodeURIComponent(key)}/artifacts/${kind}/${encodeURIComponent(id)}`, { method: 'PATCH', body: input });
  },
  deleteArtifact(key: string, kind: ArtifactKind, id: ID) {
    return apiClient.delete<void>(`/work-items/${encodeURIComponent(key)}/artifacts/${kind}/${encodeURIComponent(id)}`);
  },
};
