import { describe, expect, it } from 'vitest';
import { parseWorkItemRoute } from './workItemRoute';

describe('work item deep link', () => {
  it('restores both the project context and selected work item key', () => {
    expect(parseWorkItemRoute('/projects/project%2Falpha/work-items/FLOW-42/')).toEqual({ projectId: 'project/alpha', key: 'FLOW-42' });
  });

  it('does not claim unrelated or malformed URLs', () => {
    expect(parseWorkItemRoute('/kanban/')).toBeNull();
    expect(parseWorkItemRoute('/projects/%E0%A4%A/work-items/FLOW-42/')).toBeNull();
  });
});
