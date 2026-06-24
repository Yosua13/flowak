/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

export type ID = string;

export type NodeType = 'terminator' | 'process' | 'decision' | 'actor' | 'system';

export type RoleKey = 'uiux' | 'frontend' | 'backend';

export type Status = 'planned' | 'in_progress' | 'review' | 'done';

export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';

export interface BusinessFacet {
  actor?: string;
  trigger?: string;
  input?: string;
  process?: string;
  output?: string;
  rules?: string;
  exceptionPath?: string;
  system?: string;
  sla?: string;
  priority?: string;
  riskLevel?: string;
  acceptanceCriteria?: string;
}

export interface UiuxFacet {
  assignee?: string;
  status?: Status;
  screen?: string;
  link?: string;
  wireframeUrl?: string;
  stateNotes?: string;
  accessibilityNotes?: string;
  dueDate?: string;
  notes?: string;
}

export interface FrontendFacet {
  assignee?: string;
  status?: Status;
  page?: string;
  route?: string;
  interaction?: string;
  validation?: string;
  state?: string;
  handoffLink?: string;
  dueDate?: string;
  notes?: string;
}

export interface BackendFacet {
  assignee?: string;
  status?: Status;
  method?: HttpMethod;
  endpoint?: string;
  auth?: string;
  request?: string; // JSON text
  response?: string; // JSON text
  statusCode?: string;
  errorCodes?: string; // JSON text
  dueDate?: string;
  notes?: string;
}

export interface Node {
  id: ID;
  type: NodeType;
  label: string;
  x: number;
  y: number;
  doc: BusinessFacet;
  roles: {
    uiux?: UiuxFacet;
    frontend?: FrontendFacet;
    backend?: BackendFacet;
  };
}

export interface Edge {
  id: ID;
  from: ID;
  to: ID;
  label?: string;
}

export interface Module {
  id: ID;
  name: string;
  description?: string;
  nodes: Node[];
  edges: Edge[];
  schemaVersion: number;
}
