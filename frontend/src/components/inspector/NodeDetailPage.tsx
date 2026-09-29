import { useEffect } from 'react';
import type { ReactNode } from 'react';
import { AlertTriangle, Loader2 } from 'lucide-react';
import { useStore } from '../../store/useStore';
import type { NodeDetailRoute } from './nodeDetailRoute';
import NodeDetailSurface from './NodeDetailSurface';

interface NodeDetailPageProps {
  route: NodeDetailRoute;
  routeState: 'loading' | 'unavailable' | 'ready';
  onClose: () => void;
}

/** Full-page host for the same surface used by the canvas drawer. */
export default function NodeDetailPage({ route, routeState, onClose }: NodeDetailPageProps) {
  const { activeProjectId, modules, selectModule, selectNode } = useStore();
  const module = modules.find((candidate) => candidate.id === route.moduleId);
  const node = module?.nodes.find((candidate) => candidate.id === route.nodeId);

  useEffect(() => {
    if (routeState === 'ready' && activeProjectId === route.projectId && module && node) {
      selectModule(module.id);
      selectNode(node.id);
    }
  }, [activeProjectId, module, node, route.projectId, routeState, selectModule, selectNode]);

  if (routeState === 'loading' || activeProjectId !== route.projectId) return <PageState icon={<Loader2 className="h-5 w-5 animate-spin" />} title="Memuat detail node" message="Menyiapkan konteks proyek dan node…" />;
  if (routeState === 'unavailable') return <PageState icon={<AlertTriangle className="h-5 w-5" />} title="Detail tidak dapat dibuka" message="Anda tidak memiliki akses ke proyek ini, atau proyek sudah tidak tersedia." action={onClose} />;
  if (!module || !node) return <PageState icon={<AlertTriangle className="h-5 w-5" />} title="Node telah dihapus" message="URL ini mengarah ke modul atau node yang sudah tidak tersedia." action={onClose} />;

  return <NodeDetailSurface node={node} module={module} mode="page" onModeChange={() => undefined} onOpenFullPage={() => undefined} onClose={onClose} />;
}

function PageState({ icon, title, message, action }: { icon: ReactNode; title: string; message: string; action?: () => void }) {
  return <main className="flex min-h-screen items-center justify-center bg-[#0A0A0B] p-6 text-gray-100"><section className="w-full max-w-lg rounded-2xl border border-white/10 bg-[#131315] p-6" role="status"><div className="text-[#E2C392]">{icon}</div><h1 className="mt-3 text-lg font-bold">{title}</h1><p className="mt-2 text-sm text-gray-400">{message}</p>{action && <button onClick={action} className="mt-5 rounded-lg bg-[#C5A267] px-4 py-2 text-sm font-semibold text-black">Kembali ke kanvas</button>}</section></main>;
}
