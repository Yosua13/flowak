export interface NodeDetailRoute {
  projectId: string;
  moduleId: string;
  nodeId: string;
}

const nodeDetailPattern = /^\/projects\/([^/]+)\/modules\/([^/]+)\/nodes\/([^/]+)\/?$/;

function decodeSegment(value: string) {
  try {
    return decodeURIComponent(value);
  } catch {
    return null;
  }
}

/** Parses only the canonical, shareable full-page node detail URL. */
export function parseNodeDetailRoute(pathname: string): NodeDetailRoute | null {
  const match = pathname.match(nodeDetailPattern);
  if (!match) return null;
  const [projectId, moduleId, nodeId] = match.slice(1).map(decodeSegment);
  return projectId && moduleId && nodeId ? { projectId, moduleId, nodeId } : null;
}

export function nodeDetailPath({ projectId, moduleId, nodeId }: NodeDetailRoute) {
  return `/projects/${encodeURIComponent(projectId)}/modules/${encodeURIComponent(moduleId)}/nodes/${encodeURIComponent(nodeId)}/`;
}

export function nodeCanvasPath({ projectId, moduleId }: NodeDetailRoute) {
  const query = new URLSearchParams({ project: projectId, module: moduleId });
  return `/kanvas/?${query.toString()}`;
}

/** Returns the next index for a cyclic modal/drawer focus trap. */
export function nextFocusIndex(currentIndex: number, itemCount: number, backwards: boolean) {
  if (itemCount < 1) return -1;
  if (currentIndex < 0) return backwards ? itemCount - 1 : 0;
  return backwards ? (currentIndex - 1 + itemCount) % itemCount : (currentIndex + 1) % itemCount;
}
