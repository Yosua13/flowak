import { useState } from 'react';
import { useStore } from '../../store/useStore';
import NodeDetailSurface from './NodeDetailSurface';

/** Hosts the reusable node detail surface in the Canvas drawer. */
export default function Inspector() {
  const { modules, activeId, selectedNodeId, selectNode } = useStore();
  const [mode, setMode] = useState<'drawer' | 'modal' | 'page'>('drawer');
  const module = modules.find((candidate) => candidate.id === activeId);
  const node = module?.nodes.find((candidate) => candidate.id === selectedNodeId);

  if (!module || !node) return null;

  return <NodeDetailSurface node={node} module={module} mode={mode} onModeChange={setMode} onClose={() => selectNode(null)} />;
}
