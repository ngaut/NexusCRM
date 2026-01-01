import React from 'react';
import { NavLink } from 'react-router-dom';
import {
    Settings, FileCode, ExternalLink, ChevronLeft, ChevronRight, Inbox, Trash2
} from 'lucide-react';
import * as Icons from 'lucide-react';
import { useApp } from '../../contexts/AppContext';
import { usePendingApprovals } from '../../core/hooks/usePendingApprovals';
import { Tooltip } from '../ui/Tooltip';

interface SidebarProps {
    isSidebarCollapsed: boolean;
    setIsSidebarCollapsed: (collapsed: boolean) => void;
    mobileMenuOpen: boolean;
    setMobileMenuOpen: (open: boolean) => void;
}

export const Sidebar: React.FC<SidebarProps> = ({
    isSidebarCollapsed,
    setIsSidebarCollapsed,
    mobileMenuOpen,
    setMobileMenuOpen,
}) => {
    const { currentApp } = useApp();
    const { count: pendingApprovalsCount } = usePendingApprovals();

    // Get icon component by name
    const getIconComponent = (iconName: string) => {
        const IconComponent = Icons[iconName as keyof typeof Icons] as unknown as React.ComponentType<{ size?: number; className?: string }>;
        return IconComponent || FileCode;
    };

    const hasNavigationItems = currentApp?.navigation_items && currentApp.navigation_items.length > 0;
    const itemsToRender = currentApp?.navigation_items || [];
    const showLabels = !isSidebarCollapsed || mobileMenuOpen;

    return (
        <>
            {/* Mobile Backdrop */}
            {mobileMenuOpen && (
                <div
                    className="fixed inset-0 z-30 bg-black/50 lg:hidden backdrop-blur-sm transition-opacity"
                    onClick={() => setMobileMenuOpen(false)}
                />
            )}

            <aside
                className={`
            fixed inset-y-0 left-0 z-40 h-full lg:static lg:h-auto
            ${isSidebarCollapsed ? 'lg:w-20' : 'lg:w-64'}
            w-64 bg-white/50 backdrop-blur-lg border-r border-slate-200/60 
            flex flex-col transition-all duration-300 ease-in-out
            ${mobileMenuOpen ? 'translate-x-0 shadow-2xl' : '-translate-x-full lg:translate-x-0'}
          `}
            >
                {/* Scrollable Navigation Items */}
                <nav className="flex-1 overflow-y-auto p-3 space-y-1">
                    {/* Navigation Items from App */}
                    {hasNavigationItems && (
                        <>
                            {itemsToRender.map(item => {
                                const IconComponent = getIconComponent(item.icon);

                                const content = (
                                    <>
                                        <IconComponent size={20} className="shrink-0" />
                                        {showLabels && (
                                            <span className="truncate">{item.label}</span>
                                        )}
                                        {showLabels && item.type === 'web' && (
                                            <ExternalLink size={12} className="ml-auto text-slate-400 shrink-0" />
                                        )}
                                    </>
                                );

                                const commonClasses = ({ isActive }: { isActive: boolean }) =>
                                    `flex items-center gap-3 px-3 py-2.5 rounded-xl text-sm font-medium transition-all duration-200 ${isActive
                                        ? 'text-white shadow-md'
                                        : 'text-slate-600 hover:bg-white/60 hover:text-slate-900'
                                    } ${isSidebarCollapsed ? 'justify-center' : ''}`;

                                const activeStyle = ({ isActive }: { isActive: boolean }) => isActive ? {
                                    backgroundColor: 'var(--color-brand)',
                                    boxShadow: '0 4px 6px -1px color-mix(in srgb, var(--color-brand), transparent 80%)'
                                } : {};

                                // Web tabs
                                if (item.type === 'web' && item.page_url) {
                                    const link = (
                                        <a
                                            key={item.id}
                                            href={item.page_url}
                                            target="_blank"
                                            rel="noopener noreferrer"
                                            className={commonClasses({ isActive: false })}
                                        >
                                            {content}
                                        </a>
                                    );
                                    return isSidebarCollapsed ? (
                                        <Tooltip key={item.id} content={item.label} position="right">
                                            {link}
                                        </Tooltip>
                                    ) : link;
                                }

                                // Object tabs use NavLink
                                const getNavRoute = () => {
                                    if (item.type === 'dashboard' && item.dashboard_id) {
                                        return `/dashboard/${item.dashboard_id}`;
                                    }
                                    if (item.type === 'page' && item.page_url) {
                                        return item.page_url;
                                    }
                                    if (item.type === 'object' && item.object_api_name) {
                                        return `/object/${item.object_api_name}`;
                                    }
                                    return '/';
                                };

                                const navLink = (
                                    <NavLink
                                        key={item.id}
                                        to={getNavRoute()}
                                        end={item.page_url === '/'}
                                        className={commonClasses}
                                        style={activeStyle}
                                    >
                                        {content}
                                    </NavLink>
                                );

                                return isSidebarCollapsed ? (
                                    <Tooltip key={item.id} content={item.label} position="right">
                                        {navLink}
                                    </Tooltip>
                                ) : navLink;
                            })}
                        </>
                    )}
                </nav>

                {/* Fixed System Footer */}
                <div className="p-3 border-t border-slate-200/60 bg-white/30 backdrop-blur-sm">
                    {!isSidebarCollapsed && (
                        <div className="px-3 text-xs font-semibold text-slate-400 uppercase tracking-wider mb-2 animate-in fade-in duration-200">System</div>
                    )}

                    {(() => {
                        const setupLink = (
                            <NavLink
                                to="/setup"
                                className={({ isActive }) => `flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium transition-all duration-200 ${isActive ? 'text-white shadow-md' : 'text-slate-600 hover:bg-white/60 hover:text-slate-900'} ${isSidebarCollapsed ? 'justify-center' : ''}`}
                                style={({ isActive }) => isActive ? { backgroundColor: 'var(--color-brand)' } : {}}
                            >
                                <Settings size={20} className="shrink-0" />
                                {showLabels && <span>Setup</span>}
                            </NavLink>
                        );
                        return (!showLabels) ? (
                            <Tooltip content="Setup" position="right">{setupLink}</Tooltip>
                        ) : setupLink;
                    })()}

                    {(() => {
                        const approvalsLink = (
                            <NavLink
                                to="/approvals"
                                className={({ isActive }) => `flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium transition-all duration-200 ${isActive ? 'text-white shadow-md' : 'text-slate-600 hover:bg-white/60 hover:text-slate-900'} ${isSidebarCollapsed ? 'justify-center' : ''}`}
                                style={({ isActive }) => isActive ? { backgroundColor: 'var(--color-brand)' } : {}}
                            >
                                <div className="relative">
                                    <Inbox size={20} className="shrink-0" />
                                    {pendingApprovalsCount > 0 && (
                                        <span className="absolute -top-1.5 -right-1.5 w-4 h-4 bg-red-500 text-white text-[10px] font-bold rounded-full flex items-center justify-center">
                                            {pendingApprovalsCount > 9 ? '9+' : pendingApprovalsCount}
                                        </span>
                                    )}
                                </div>
                                {showLabels && (
                                    <span className="flex items-center gap-2">
                                        Approvals
                                        {pendingApprovalsCount > 0 && (
                                            <span className="px-1.5 py-0.5 bg-red-100 text-red-600 text-xs font-semibold rounded">
                                                {pendingApprovalsCount}
                                            </span>
                                        )}
                                    </span>
                                )}
                            </NavLink>
                        );
                        return (!showLabels) ? (
                            <Tooltip content="Approvals" position="right">{approvalsLink}</Tooltip>
                        ) : approvalsLink;
                    })()}

                    {(() => {
                        const recycleLink = (
                            <NavLink
                                to="/recyclebin"
                                className={({ isActive }) => `flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium transition-all duration-200 ${isActive ? 'text-white shadow-md' : 'text-slate-600 hover:bg-white/60 hover:text-slate-900'} ${isSidebarCollapsed ? 'justify-center' : ''}`}
                                style={({ isActive }) => isActive ? { backgroundColor: 'var(--color-brand)' } : {}}
                            >
                                <Trash2 size={20} className="shrink-0" />
                                {showLabels && <span>Recycle Bin</span>}
                            </NavLink>
                        );
                        return (!showLabels) ? (
                            <Tooltip content="Recycle Bin" position="right">{recycleLink}</Tooltip>
                        ) : recycleLink;
                    })()}

                    {/* Collapse Toggle - Desktop Only */}
                    <button
                        onClick={() => setIsSidebarCollapsed(!isSidebarCollapsed)}
                        className={`hidden lg:flex mt-2 w-full items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium text-slate-400 hover:bg-white/60 hover:text-slate-700 transition-all ${isSidebarCollapsed ? 'justify-center' : ''}`}
                        title={isSidebarCollapsed ? "Expand Sidebar" : "Collapse Sidebar"}
                    >
                        {isSidebarCollapsed ? <ChevronRight size={20} /> : <ChevronLeft size={20} />}
                        {!isSidebarCollapsed && <span>Collapse</span>}
                    </button>
                </div>
            </aside>
        </>
    );
};
