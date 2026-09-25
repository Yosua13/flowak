import { useState } from 'react';
import { useStore } from '../../store/useStore';
import NodeDetailSurface from './NodeDetailSurface';

/** Hosts the reusable node detail surface in the Canvas drawer. */
export default function Inspector({ onOpenFullPage }: { onOpenFullPage: () => void }) {
  const { modules, activeId, selectedNodeId, selectNode } = useStore();
  const [mode, setMode] = useState<'drawer' | 'modal'>('drawer');
  const module = modules.find((candidate) => candidate.id === activeId);
  const node = module?.nodes.find((candidate) => candidate.id === selectedNodeId);

  if (!module || !node) return null;

  return <NodeDetailSurface node={node} module={module} mode={mode} onModeChange={setMode} onOpenFullPage={onOpenFullPage} onClose={() => selectNode(null)} />;
}
