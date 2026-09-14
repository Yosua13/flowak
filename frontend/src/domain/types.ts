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
  outcome?: string;
  actor?: string;
  triggerType?: 'manual' | 'event' | 'schedule' | 'api';
  trigger?: string;
  preconditions?: string;
  input?: string;
  process?: string;
  output?: string;
  rules?: BusinessRule[];
  decisionOutcomes?: DecisionOutcome[];
  exceptionPaths?: string;
  exceptionPath?: string;
  system?: string;
  sla?: string;
  priority?: string;
  riskLevel?: string;
  acceptanceCriteria?: string;
  referenceLinks?: string;
}

export interface BusinessRule { code?: string; severity?: 'low' | 'medium' | 'high' | 'critical'; description: string; }
export interface DecisionOutcome { outcome: string; edgeId?: ID; }

export interface UiuxFacet {
  assignee?: string;
  reviewer?: string;
  readiness?: Status;
  status?: Status;
  userGoal?: string;
  surface?: string;
  screen?: string;
  figmaFrameUrl?: string;
  designVersion?: string;
  screenStates?: string;
  interactions?: string;
  contentMessages?: string;
  responsiveIntent?: string;
  link?: string;
  wireframeUrl?: string;
  stateNotes?: string;
  accessibilityNotes?: string;
  dueDate?: string;
  notes?: string;
}

export interface FrontendFacet {
  assignee?: string;
  reviewer?: string;
  readiness?: Status;
  status?: Status;
  experienceName?: string;
  page?: string;
  route?: string;
  entryExitBehavior?: string;
  interaction?: string;
  inputRequirements?: string;
  validation?: string;
  state?: string;
  apiReferences?: string;
  analyticsIntent?: string;
  featureAvailability?: string;
  handoffLink?: string;
  dueDate?: string;
  notes?: string;
}

export interface BackendFacet {
  assignee?: string;
  reviewer?: string;
  readiness?: Status;
  status?: Status;
  serviceCapability?: string;
  apiReferences?: string;
  method?: HttpMethod;
  endpoint?: string;
  auth?: string;
  request?: string; // JSON text
  response?: string; // JSON text
  statusCode?: string;
  errorCodes?: string; // JSON text
  businessValidation?: string;
  dependencyReferences?: string;
  idempotencyNotes?: string;
  cachingNotes?: string;
  securityNotes?: string;
  observabilityIntent?: string;
  sla?: string;
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
  rowVersion?: number;
}

export interface Edge {
  id: ID;
  from: ID;
  to: ID;
  label?: string;
  rowVersion?: number;
}

export interface GraphDelete {
  id: ID;
  rowVersion: number;
}

export type WorkItemType = 'Story' | 'Task' | 'Bug' | 'Review' | 'Research' | 'Subtask';
export type WorkItemStatus = 'Backlog' | 'Ready' | 'In Progress' | 'In Review' | 'Blocked' | 'Done' | 'Canceled';

export interface WorkItem {
  id: ID;
  key: string;
  project_id: ID;
  module_id?: ID;
  node_id?: ID;
  facet_key?: string;
  parent_id?: ID;
  type: WorkItemType;
  title: string;
  description?: string;
  priority: 'low' | 'medium' | 'high' | 'critical';
  points?: 1 | 2 | 3 | 5 | 8 | 13;
  status: WorkItemStatus;
  assignee_id?: ID;
  reporter_id: ID;
  row_version: number;
}

export interface Module {
  id: ID;
  name: string;
  description?: string;
  nodes: Node[];
  edges: Edge[];
  schemaVersion: number;
  rowVersion?: number;
  deletedNodes?: GraphDelete[];
  deletedEdges?: GraphDelete[];
}
