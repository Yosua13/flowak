/**
 * @license
 * SPDX-License-Identifier: Apache-2.0
 */

import React, { useEffect } from 'react';
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
import Login from './auth/Login';
import Register from './auth/Register';
import ProjectHub from './dashboard/ProjectHub';
import { Loader2, X } from 'lucide-react';
import { motion, AnimatePresence } from 'motion/react';

export default function AppShell() {
  const { screen, view, selectedNodeId, initializeStore, selectedNotif, setSelectedNotif } = useStore();

  useEffect(() => {
    initializeStore();
  }, [initializeStore]);

  useEffect(() => {
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
    const newUrl = `${window.location.origin}${path}`;
    window.history.pushState({ screen, view }, '', newUrl);
  }, [screen, view]);

  // Routing based on screen state
  if (screen === 'login') {
    return <Login />;
  }

  if (screen === 'register') {
    return <Register />;
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
        <div className="print:hidden">
          <Topbar />
        </div>
        <div className="flex-1 h-full overflow-hidden flex flex-col relative print:overflow-visible print:h-auto">
          {renderActiveLens()}
        </div>
      </div>

      {/* Inspector - right zone (reveals only if a node is focused in Canvas) */}
      {selectedNodeId && (
        <div className="print:hidden">
          <Inspector />
        </div>
      )}

      {/* Global Notification Detail Modal */}
      <AnimatePresence>
        {selectedNotif && (
          <div className="fixed inset-0 bg-black/60 backdrop-blur-sm z-[9999] flex items-center justify-center p-4">
            <motion.div
              initial={{ opacity: 0, scale: 0.95 }}
              animate={{ opacity: 1, scale: 1 }}
              exit={{ opacity: 0, scale: 0.95 }}
              className="bg-[#131315] border border-white/10 w-full max-w-md rounded-2xl p-6 shadow-2xl space-y-4"
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
          </div>
        )}
      </AnimatePresence>

    </div>
  );
}
