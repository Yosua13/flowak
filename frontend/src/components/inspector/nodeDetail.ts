import type { Edge, Module, Node, Status, WorkItem } from '../../domain/types';

export type DetailTab = 'summary' | 'business' | 'uiux' | 'frontend' | 'backend' | 'tasks' | 'discussion' | 'activity';

const facetStates: Status[] = ['planned', 'in_progress', 'review', 'done'];

export function facetReadiness(node: Node, facet: 'uiux' | 'frontend' | 'backend'): Status | undefined {
  const value = node.roles[facet];
  return value?.readiness || value?.status;
}

export function nodeCompleteness(node: Node): number {
  const checks = [
    node.label,
    node.doc.outcome || node.doc.output,
    node.doc.actor,
    node.doc.trigger,
    node.doc.process,
    ...(['uiux', 'frontend', 'backend'] as const).map((facet) => facetReadiness(node, facet)),
  ];
  return Math.round((checks.filter(Boolean).length / checks.length) * 100);
}

export function nodeRisk(node: Node): string {
  return node.doc.riskLevel || 'not set';
}

export function nextPaths(node: Node, edges: Edge[], nodes: Node[]): Array<{ id: string; label: string; edgeLabel?: string }> {
  return edges
    .filter((edge) => edge.from === node.id)
    .map((edge) => ({ id: edge.to, label: nodes.find((candidate) => candidate.id === edge.to)?.label || 'Node tidak ditemukan', edgeLabel: edge.label }));
}

export function nodeWorkItems(items: WorkItem[], nodeId: string): WorkItem[] {
  return items.filter((item) => item.node_id === nodeId);
}

export function isOpenWorkItem(item: WorkItem): boolean {
  return item.status !== 'Done' && item.status !== 'Canceled';
}

export function isOverdue(item: WorkItem, today = new Date()): boolean {
  const dueDate = item.due_date;
  if (!dueDate || !isOpenWorkItem(item)) return false;
  return new Date(`${dueDate}T23:59:59`).getTime() < today.getTime();
}

export function detailTabs(): Array<{ id: DetailTab; label: string }> {
  return [
    { id: 'summary', label: 'Ringkasan' },
    { id: 'business', label: 'Bisnis' },
    { id: 'uiux', label: 'UI/UX' },
    { id: 'frontend', label: 'Frontend' },
    { id: 'backend', label: 'Backend' },
    { id: 'tasks', label: 'Task' },
    { id: 'discussion', label: 'Diskusi' },
    { id: 'activity', label: 'Aktivitas' },
  ];
}

export function isKnownFacetState(value: Status | undefined): boolean {
  return value !== undefined && facetStates.includes(value);
}

export function moduleForNode(modules: Module[], moduleId: string | null): Module | undefined {
  return modules.find((module) => module.id === moduleId);
}
