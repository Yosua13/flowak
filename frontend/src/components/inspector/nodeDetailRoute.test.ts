import { describe, expect, it } from 'vitest';
import { nextFocusIndex, nodeCanvasPath, nodeDetailPath, parseNodeDetailRoute } from './nodeDetailRoute';

describe('node detail route and keyboard boundary helpers', () => {
  it('round-trips an encoded project, module, and node through the full-page URL', () => {
    const route = { projectId: 'project/team a', moduleId: 'module-1', nodeId: 'node/a' };
    expect(parseNodeDetailRoute(nodeDetailPath(route))).toEqual(route);
    expect(nodeCanvasPath(route)).toBe('/kanvas/?project=project%2Fteam+a&module=module-1');
  });

  it('rejects non-canonical and malformed deep links', () => {
    expect(parseNodeDetailRoute('/projects/p/nodes/n/')).toBeNull();
    expect(parseNodeDetailRoute('/projects/%E0%A4%A/modules/m/nodes/n/')).toBeNull();
  });

  it('cycles focus at both ends of a drawer or modal', () => {
    expect(nextFocusIndex(2, 3, false)).toBe(0);
    expect(nextFocusIndex(0, 3, true)).toBe(2);
    expect(nextFocusIndex(-1, 3, false)).toBe(0);
  });
});
