import { describe, expect, it } from 'vitest';
import type { Module, WorkItem } from '../domain/types';
import { buildCalendarEvents, type DerivedViewData } from './derivedViewData';
import { calculateDerivedMetrics } from './derivedMetrics';
import { generateCanonicalExport, generateMarkdown, generateOpenApi } from './exportService';

const module: Module = {
  id: 'module-1', name: 'Checkout', schemaVersion: 7, edges: [],
  nodes: [{ id: 'node-1', type: 'process', label: 'Charge payment', x: 0, y: 0, doc: {}, roles: { uiux: { readiness: 'done' }, frontend: { readiness: 'done' }, backend: { readiness: 'review', method: 'POST', endpoint: '/payments?api_key=secret', auth: 'Bearer secret', request: '{"token":"request-secret"}', response: '{"token":"response-secret"}' } } }],
};

const workItem: WorkItem = { id: 'work-1', key: 'FLOW-1', project_id: 'project-1', module_id: 'module-1', node_id: 'node-1', facet_key: 'backend', type: 'Task', title: 'Ship payment', priority: 'high', status: 'Done', reporter_id: 'user-1', row_version: 1, created_at: '2026-01-01T00:00:00Z', start_date: '2026-01-01', due_date: '2026-01-02' };

const fixture: DerivedViewData = {
  work_items: [workItem],
  history: [
    { work_item_id: 'work-1', to_status: 'Backlog', created_at: '2026-01-01T00:00:00Z' },
    { work_item_id: 'work-1', from_status: 'Backlog', to_status: 'Blocked', created_at: '2026-01-01T02:00:00Z' },
    { work_item_id: 'work-1', from_status: 'Blocked', to_status: 'In Progress', created_at: '2026-01-01T05:00:00Z' },
    { work_item_id: 'work-1', from_status: 'In Progress', to_status: 'Done', created_at: '2026-01-01T10:00:00Z' },
  ],
  facet_reviews: [{ node_id: 'node-1', node_label: 'Charge payment', role_key: 'backend', readiness: 'review', status: 'review', due_date: '2026-01-03' }],
  baselines: [{ module_id: 'module-1', version: 4, created_at: '2026-01-04T00:00:00Z' }],
  comments: [{ id: 'comment-1', work_item_id: 'work-1', body: 'Decision recorded', created_at: '2026-01-01T01:00:00Z' }],
  evidence: [{ id: 'evidence-1', kind: 'attachment', label: 'payment-proof.pdf', work_item_id: 'work-1', node_id: 'node-1', created_at: '2026-01-01T11:00:00Z' }],
  generated_at: '2026-01-05T00:00:00Z',
};

describe('derived view fixture contract', () => {
  it('projects the same normalized records into calendar, metric, and document views', () => {
    const events = buildCalendarEvents(fixture);
    const metrics = calculateDerivedMetrics(fixture.work_items, fixture.history, module, new Date('2026-01-05T00:00:00Z'));
    const document = generateMarkdown(module, fixture);

    expect(events).toEqual(expect.arrayContaining([
      expect.objectContaining({ id: 'work-1-start', title: 'Ship payment' }),
      expect.objectContaining({ id: 'work-1-due', detail: 'FLOW-1 · Done' }),
      expect.objectContaining({ id: 'node-1-backend', title: 'Charge payment' }),
      expect.objectContaining({ id: 'module-1-4', title: 'Baseline v4' }),
    ]));
    expect(metrics).toMatchObject({ throughput: 1, cycleTimeHours: 10, blockedTimeHours: 3, traceability: 1 });
    expect(document).toContain('FLOW-1');
    expect(document).toContain('Decision recorded');
    expect(document).toContain('Baseline modul**: v4');
  });

  it('preserves schema version and redacts request and response secrets in every export', () => {
    const canonical = generateCanonicalExport(module, fixture.generated_at);
    const openapi = generateOpenApi(module);
    expect(JSON.parse(canonical).schemaVersion).toBe(7);
    expect(JSON.parse(openapi).openapi).toBe('3.1.0');
    [canonical, openapi].forEach((output) => {
      expect(output).not.toContain('secret');
      expect(output).not.toContain('response-secret');
    });
  });
});
