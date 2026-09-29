export interface WorkItemRoute { projectId: string; key: string; }

export function parseWorkItemRoute(pathname: string): WorkItemRoute | null {
  const match = pathname.match(/^\/projects\/([^/]+)\/work-items\/([^/]+)\/?$/);
  if (!match) return null;
  try {
    return { projectId: decodeURIComponent(match[1]), key: decodeURIComponent(match[2]) };
  } catch {
    return null;
  }
}
