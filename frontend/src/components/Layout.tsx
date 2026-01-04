import React, { ReactNode, useState, useEffect } from 'react';
import { useLocation } from 'react-router-dom';
import { useApp } from '../contexts/AppContext';
import { UtilityBar } from './UtilityBar';
import { TopBar } from './layout/TopBar';
import { Sidebar } from './layout/Sidebar';
import { STORAGE_KEYS } from '../core/constants/ApplicationDefaults';

interface LayoutProps {
  children: ReactNode;
  onToggleAI: () => void;
}

export function Layout({ children, onToggleAI }: LayoutProps) {
  const { currentApp } = useApp();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  // Sidebar collapsed state with persistence
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(() => {
    return localStorage.getItem(STORAGE_KEYS.SIDEBAR_COLLAPSED) === 'true';
  });

  useEffect(() => {
    localStorage.setItem(STORAGE_KEYS.SIDEBAR_COLLAPSED, String(isSidebarCollapsed));
  }, [isSidebarCollapsed]);

  // Close mobile menu on route change
  const location = useLocation();
  useEffect(() => {
    setMobileMenuOpen(false);
  }, [location]);

  return (
    <div className="h-screen flex flex-col overflow-hidden bg-slate-50 relative">

      <TopBar
        mobileMenuOpen={mobileMenuOpen}
        setMobileMenuOpen={setMobileMenuOpen}
        isSidebarCollapsed={isSidebarCollapsed}
        onToggleAI={onToggleAI}
      />

      <div className="flex flex-1 overflow-hidden relative">
        <Sidebar
          isSidebarCollapsed={isSidebarCollapsed}
          setIsSidebarCollapsed={setIsSidebarCollapsed}
          mobileMenuOpen={mobileMenuOpen}
          setMobileMenuOpen={setMobileMenuOpen}
        />

        <main className="flex-1 overflow-y-auto bg-slate-50 p-6">
          {children}
        </main>
      </div >

      {/* Utility Bar */}
      {
        currentApp?.utility_items && currentApp.utility_items.length > 0 && (
          <UtilityBar items={currentApp.utility_items} />
        )
      }
    </div >
  );
}
