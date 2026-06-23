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
import { Loader2 } from 'lucide-react';

export default function AppShell() {
  const { view, selectedNodeId, initializeStore, modules } = useStore();

  useEffect(() => {
    initializeStore();
  }, [initializeStore]);

  // If initial store is empty and still initializing
  if (modules.length === 0) {
    return (
      <div className="h-screen w-screen flex flex-col items-center justify-center bg-[#0A0A0B] text-white">
        <Loader2 className="w-8 h-8 text-[#C5A267] animate-spin mb-4" />
        <p className="text-xs font-mono text-gray-400">Menyiapkan Workspace Flowak...</p>
      </div>
    );
  }

  // Selective rendering of active lens of model
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

    </div>
  );
}
