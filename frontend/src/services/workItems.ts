import type { ID, WorkItem, WorkItemStatus, WorkItemType } from '../domain/types';

export interface WorkItemInput {
  module_id?: ID; node_id?: ID; facet_key?: string; parent_id?: ID;
  type: WorkItemType; title: string; description?: string;
  priority?: WorkItem['priority']; points?: WorkItem['points']; assignee_id?: ID;
}

/** API contract only; state remains server-authoritative. */
export const workItemsApi = {
  list(projectId: ID, query = ''): Promise<WorkItem[]> {
    return fetch(`/api/projects/${projectId}/work-items${query ? `?${query}` : ''}`).then(r => r.json());
  },
  create(projectId: ID, input: WorkItemInput): Promise<WorkItem> {
    return fetch(`/api/projects/${projectId}/work-items`, { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify(input) }).then(r => r.json());
  },
  transition(key: string, status: WorkItemStatus, row_version: number, note = ''): Promise<WorkItem> {
    return fetch(`/api/work-items/${key}/transitions`, { method: 'POST', headers: {'Content-Type': 'application/json'}, body: JSON.stringify({status, row_version, note}) }).then(r => r.json());
  },
};
