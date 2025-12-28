import React from 'react';
import { Routes, Route, Navigate, Outlet } from 'react-router-dom';
import { Layout } from './Layout';
import { Dashboard } from '../pages/Dashboard';
import { ObjectView } from '../pages/ObjectView';
import { Setup } from '../pages/Setup';
import { AppStudio } from '../pages/AppStudio';
import { RecycleBin } from '../pages/RecycleBin';
import { ApprovalQueue } from '../pages/ApprovalQueue';

interface RouterProps {
  onToggleAI: () => void;
}

export function Router({ onToggleAI }: RouterProps) {
  return (
    <Routes>
      {/* App Studio - Standalone Layout */}
      <Route path="/studio/:appId" element={<AppStudio />} />

      {/* Standard App Layout */}
      <Route
        element={
          <Layout onToggleAI={onToggleAI}>
            <Outlet />
          </Layout>
        }
      >
        <Route path="/" element={<Dashboard />} />
        {/* Dashboard viewing - redirects to library (consolidated to App-based) */}
        <Route path="/dashboards" element={<Dashboard />} />
        <Route path="/dashboard/:dashboardId" element={<Dashboard />} />
        <Route path="/object/:objectApiName" element={<ObjectView />} />
        <Route path="/object/:objectApiName/:recordId" element={<ObjectView />} />
        <Route path="/setup/*" element={<Setup />} />
        <Route path="/recyclebin" element={<RecycleBin />} />
        <Route path="/approvals" element={<ApprovalQueue />} />
      </Route>

      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
