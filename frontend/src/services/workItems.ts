import type { ID, WorkItem, WorkItemStatus, WorkItemType } from '../domain/types';
import { apiClient } from './apiClient';

export interface WorkItemInput {
  module_id?: ID; node_id?: ID; facet_key?: string; parent_id?: ID;
  type: WorkItemType; title: string; description?: string;
  priority?: WorkItem['priority']; points?: WorkItem['points']; assignee_id?: ID;
  status?: WorkItemStatus; start_date?: string; due_date?: string; blocked_reason?: string;
}

/** API contract only; state remains server-authoritative. */
export const workItemsApi = {
  list(projectId: ID, query = ''): Promise<WorkItem[]> {
    return apiClient.get(`/projects/${projectId}/work-items${query ? `?${query}` : ''}`);
  },
  create(projectId: ID, input: WorkItemInput): Promise<WorkItem> {
    return apiClient.post(`/projects/${projectId}/work-items`, input);
  },
  transition(key: string, status: WorkItemStatus, row_version: number, note = ''): Promise<WorkItem> {
    const resolution = status === 'Done' ? note : undefined;
    return apiClient.post(`/work-items/${key}/transitions`, { status, row_version, note, resolution });
  },
  update(key: string, input: Partial<WorkItem> & { row_version: number }): Promise<WorkItem> {
    return apiClient.request(`/work-items/${key}`, { method: 'PATCH', body: input });
  },
  comments(key: string) {
    return apiClient.get<Array<{ id: ID; parent_id?: ID; author_id: ID; body: string; resolved_at?: string; created_at: string; updated_at: string }>>(`/work-items/${key}/comments`);
  },
  comment(key: string, body: string) {
    return apiClient.post<{ id: ID; body: string }>(`/work-items/${key}/comments`, { body });
  },
};
