import { renderToStaticMarkup } from 'react-dom/server';
import { describe, expect, it } from 'vitest';
import type { Module, Node } from '../../domain/types';
import NodeDetailSurface from './NodeDetailSurface';

const node: Node = { id: 'node-1', type: 'process', label: 'Verifikasi pesanan', x: 0, y: 0, doc: {}, roles: {} };
const module: Module = { id: 'module-1', name: 'Alur utama', nodes: [node], edges: [], schemaVersion: 1 };
const sharedProps = { node, module, onClose: () => undefined, onModeChange: () => undefined, onOpenFullPage: () => undefined };

describe('NodeDetailSurface accessibility hosts', () => {
  it('uses a modal dialog for drawer detail and a main landmark for the shareable page', () => {
    const drawer = renderToStaticMarkup(<NodeDetailSurface {...sharedProps} mode="drawer" />);
    const page = renderToStaticMarkup(<NodeDetailSurface {...sharedProps} mode="page" />);

    expect(drawer).toContain('role="dialog"');
    expect(drawer).toContain('aria-modal="true"');
    expect(page).toContain('role="main"');
    expect(page).not.toContain('aria-modal="true"');
  });
});
