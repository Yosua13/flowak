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
import { Loader2 } from 'lucide-react';

export default function AppShell() {
  const { screen, view, selectedNodeId, initializeStore } = useStore();

  useEffect(() => {
    initializeStore();
  }, [initializeStore]);

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

    </div>
  );
}
