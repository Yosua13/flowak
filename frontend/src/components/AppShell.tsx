/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useCallback, useEffect, useState } from 'react';
import { useStore } from '../store/useStore';
import Sidebar from './Sidebar';
import Topbar from './Topbar';
import Canvas from './canvas/Canvas';
import StatusView from './status/StatusView';
import DocView from './doc/DocView';
import CalendarView from './calendar/CalendarView';
import AnalyticsView from './analytics/AnalyticsView';
import KanbanView from './kanban/KanbanView';
import TeamView from './team/TeamView';
import Inspector from './inspector/Inspector';
import NodeDetailPage from './inspector/NodeDetailPage';
import { nodeCanvasPath, nodeDetailPath, parseNodeDetailRoute, type NodeDetailRoute } from './inspector/nodeDetailRoute';
import Login from './auth/Login';
import Register from './auth/Register';
import ProjectHub from './dashboard/ProjectHub';
import { Loader2, X } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';

export default function AppShell() {
  const { screen, view, activeProjectId, activeId, selectedNodeId, selectedWorkItemKey, initializeStore, selectProject, selectModule, selectNode, selectWorkItem, setView, selectedNotif, setSelectedNotif } = useStore();
  const [nodeRoute, setNodeRoute] = useState<NodeDetailRoute | null>(() => parseNodeDetailRoute(window.location.pathname));
  const [nodeRouteState, setNodeRouteState] = useState<'loading' | 'unavailable' | 'ready'>(() => nodeRoute ? 'loading' : 'ready');

  const loadNodeRoute = useCallback(async (route: NodeDetailRoute) => {
    setNodeRouteState('loading');
    await selectProject(route.projectId);
    if (useStore.getState().activeProjectId !== route.projectId) {
      setNodeRouteState('unavailable');
      return;
    }
    selectModule(route.moduleId);
    selectNode(route.nodeId);
    setView('canvas');
    setNodeRouteState('ready');
  }, [selectModule, selectNode, selectProject, setView]);

  useEffect(() => {
    const restoreSharedDetail = async () => {
      const params = new URLSearchParams(window.location.search);
      const detailRoute = parseNodeDetailRoute(window.location.pathname);
      const workItemPath = window.location.pathname.match(/^\/projects\/([^/]+)\/work-items\/([^/]+)\/?$/);
      await initializeStore();
      if (detailRoute) {
        setNodeRoute(detailRoute);
        await loadNodeRoute(detailRoute);
        return;
      }
      const projectId = workItemPath ? decodeURIComponent(workItemPath[1]) : params.get('project');
      const moduleId = params.get('module');
      const nodeId = params.get('node');
      const workItemKey = workItemPath ? decodeURIComponent(workItemPath[2]) : params.get('workItem');
      if (!projectId) return;
      await selectProject(projectId);
      if (moduleId) selectModule(moduleId);
      if (nodeId) selectNode(nodeId);
      if (workItemKey) {
        setView('kanban');
        selectWorkItem(workItemKey);
      }
    };
    void restoreSharedDetail();
  }, [initializeStore, loadNodeRoute, selectModule, selectNode, selectProject, selectWorkItem, setView]);

  useEffect(() => {
    const restoreFromHistory = () => {
      const route = parseNodeDetailRoute(window.location.pathname);
      setNodeRoute(route);
      if (route) void loadNodeRoute(route);
    };
    window.addEventListener('popstate', restoreFromHistory);
    return () => window.removeEventListener('popstate', restoreFromHistory);
  }, [loadNodeRoute]);

  useEffect(() => {
    if (nodeRoute) return;
    let path = '/';
    if (screen === 'login') {
      path = '/login/';
    } else if (screen === 'register') {
      path = '/register/';
    } else if (screen === 'dashboard') {
      path = '/dashboard/';
    } else if (screen === 'workspace') {
      switch (view) {
        case 'canvas':
          path = '/kanvas/';
          break;
        case 'status':
          path = '/status/';
          break;
        case 'doc':
          path = '/spesifikasi/';
          break;
        case 'calendar':
          path = '/jadwal/';
          break;
        case 'analytics':
          path = '/analitik/';
          break;
        case 'kanban':
          path = '/kanban/';
          break;
        case 'team':
          path = '/tim/';
          break;
      }
    }
    const params = new URLSearchParams();
    if (selectedNodeId && activeProjectId && activeId) {
      params.set('project', activeProjectId);
      params.set('module', activeId);
      params.set('node', selectedNodeId);
    }
    if (selectedWorkItemKey && activeProjectId) {
      path = `/projects/${encodeURIComponent(activeProjectId)}/work-items/${encodeURIComponent(selectedWorkItemKey)}/`;
      params.delete('project');
      params.delete('workItem');
    }
    const newUrl = `${window.location.origin}${path}${params.size ? `?${params.toString()}` : ''}`;
    window.history.pushState({ screen, view }, '', newUrl);
  }, [screen, view, activeProjectId, activeId, selectedNodeId, selectedWorkItemKey, nodeRoute]);

  const openNodeDetailPage = () => {
    if (!activeProjectId || !activeId || !selectedNodeId) return;
    const route = { projectId: activeProjectId, moduleId: activeId, nodeId: selectedNodeId };
    window.history.pushState({ nodeDetail: route }, '', nodeDetailPath(route));
    setNodeRoute(route);
    setNodeRouteState('ready');
  };

  const closeNodeDetailPage = () => {
    if (!nodeRoute) return;
    selectNode(null);
    setView('canvas');
    window.history.pushState({ view: 'canvas' }, '', nodeCanvasPath(nodeRoute));
    setNodeRoute(null);
  };

  // Routing based on screen state
  if (screen === 'login') {
    return <Login />;
  }

  if (screen === 'register') {
    return <Register />;
  }

  if (nodeRoute) {
    return <NodeDetailPage route={nodeRoute} routeState={nodeRouteState} onClose={closeNodeDetailPage} />;
  }

  if (screen === 'dashboard') {
    return <ProjectHub />;
  }

  // Selective rendering of active lens in Workspace
  const renderActiveLens = () => {
    switch (view) {
      case 'canvas':
        return <Canvas />;
      case 'status':
        return <StatusView />;
      case 'doc':
        return <DocView />;
      case 'calendar':
        return <CalendarView />;
      case 'analytics':
        return <AnalyticsView />;
      case 'kanban':
        return <KanbanView />;
      case 'team':
        return <TeamView />;
      default:
        return <Canvas />;
    }
  };

  return (
    <div className="h-screen w-screen flex overflow-hidden bg-[#0A0A0B] font-sans text-gray-100 transition-colors duration-150 antialiased select-none print:bg-white print:h-auto print:overflow-visible">
      
      {/* Sidebar - left zone (collapsible, dark styling) */}
      <div className="print:hidden">
        <Sidebar />
      </div>

      {/* Main lens area - middle zone */}
      <div className="flex-1 flex flex-col h-full overflow-hidden relative print:overflow-visible print:h-auto">
        <div className="print:hidden relative z-30">
          <Topbar />
        </div>
        <div className="flex-1 h-full overflow-hidden flex flex-col relative z-10 print:overflow-visible print:h-auto">
          {renderActiveLens()}
        </div>
      </div>

      {/* Inspector - right zone (reveals only if a node is focused in Canvas) */}
      {selectedNodeId && (
        <div className="print:hidden max-[768px]:fixed max-[768px]:inset-0 max-[768px]:z-[100]">
          <Inspector onOpenFullPage={openNodeDetailPage} />
        </div>
      )}

      {/* Global Notification Detail Modal */}
      <AnimatePresence>
        {selectedNotif && (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            onClick={() => setSelectedNotif(null)}
            className="fixed inset-0 bg-black/60 backdrop-blur-sm z-[9999] flex items-center justify-center p-4"
          >
            <motion.div
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.95 }}
              onClick={(e) => e.stopPropagation()}
              className="bg-[#131315] border border-white/10 w-full max-w-md rounded-2xl p-6 shadow-2xl space-y-4 text-left"
            >
              <div className="flex items-center justify-between border-b border-white/5 pb-3">
                <div className="flex items-center space-x-2">
                  <span className={`w-2.5 h-2.5 rounded-full ${
                    selectedNotif.type === 'success' ? 'bg-emerald-500' :
                    selectedNotif.type === 'warning' ? 'bg-rose-500' : 'bg-blue-500'
                  }`} />
                  <h3 className="text-sm font-bold text-white uppercase tracking-wider">{selectedNotif.title}</h3>
                </div>
                <button
                  onClick={() => setSelectedNotif(null)}
                  className="p-1 hover:bg-white/5 rounded-lg text-gray-400 hover:text-white transition cursor-pointer"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
              <p className="text-xs text-gray-300 leading-relaxed font-sans font-medium">{selectedNotif.message}</p>
              <div className="flex justify-between items-center pt-2 text-[10px] text-gray-500 font-mono">
                <span>Waktu: {selectedNotif.timestamp}</span>
                <button
                  onClick={() => setSelectedNotif(null)}
                  className="px-4 py-1.5 rounded-lg bg-white/5 hover:bg-white/10 text-xs font-bold text-white transition cursor-pointer"
                >
                  Tutup
                </button>
              </div>
            </motion.div>
          </motion.div>
        )}
      </AnimatePresence>

    </div>
  );
}
