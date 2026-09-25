import type { ID, NodeComment, WorkItem } from '../domain/types';
import { apiClient } from './apiClient';
import { workItemsApi, type WorkItemInput } from './workItems';

export interface NodeActivity {
  id: string;
  action: string;
  actor_id?: ID;
  created_at: string;
}

export interface NodeDetailData {
  items: WorkItem[];
  comments: NodeComment[];
  activity: NodeActivity[];
}

const detailCache = new Map<string, { expiresAt: number; data: NodeDetailData }>();
const pendingDetails = new Map<string, Promise<NodeDetailData>>();
const cacheKey = (projectId: ID, nodeId: ID) => `${projectId}:${nodeId}`;
const cacheWindowMs = 5_000;

/** A single detail-data contract shared by drawer, modal, and full page. */
export const nodeDetailsApi = {
  async load(projectId: ID, nodeId: ID): Promise<NodeDetailData> {
    const key = cacheKey(projectId, nodeId);
    const cached = detailCache.get(key);
    if (cached && cached.expiresAt > Date.now()) return cached.data;
    const inFlight = pendingDetails.get(key);
    if (inFlight) return inFlight;
    const encodedNodeID = encodeURIComponent(nodeId);
    const request = Promise.all([
      workItemsApi.list(projectId, `node_id=${encodedNodeID}`),
      apiClient.get<NodeComment[]>(`/nodes/${encodedNodeID}/comments`),
      apiClient.get<NodeActivity[]>(`/nodes/${encodedNodeID}/activity`),
    ]).then(([items, comments, activity]) => {
      const data = { items, comments, activity };
      detailCache.set(key, { data, expiresAt: Date.now() + cacheWindowMs });
      return data;
    }).finally(() => pendingDetails.delete(key));
    pendingDetails.set(key, request);
    return request;
  },
  async createWorkItem(projectId: ID, input: WorkItemInput) {
    const item = await workItemsApi.create(projectId, input);
    if (input.node_id) detailCache.delete(cacheKey(projectId, input.node_id));
    return item;
  },
  async createComment(nodeId: ID, body: string) {
    const comment = await apiClient.post<{ id: ID; body: string }>(`/nodes/${encodeURIComponent(nodeId)}/comments`, { body });
    for (const key of detailCache.keys()) if (key.endsWith(`:${nodeId}`)) detailCache.delete(key);
    return comment;
  },
};
