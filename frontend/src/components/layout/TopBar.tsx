import React, { useState, useRef, useEffect } from 'react';
import { Menu, LogOut, Bot } from 'lucide-react';
import { useRuntime } from '../../contexts/RuntimeContext';
import { useApp } from '../../contexts/AppContext';
import { GlobalSearch } from '../GlobalSearch';
import { NotificationCenter } from '../NotificationCenter';
import { AppLauncher } from '../AppLauncher';
import { STORAGE_KEYS } from '../../core/constants/ApplicationDefaults';

interface TopBarProps {
    mobileMenuOpen: boolean;
    setMobileMenuOpen: (open: boolean) => void;
    isSidebarCollapsed: boolean;
    onToggleAI: () => void;
}

export const TopBar: React.FC<TopBarProps> = ({
    mobileMenuOpen,
    setMobileMenuOpen,
    isSidebarCollapsed,
    onToggleAI,
}) => {
    const { user, logout } = useRuntime();
    const { currentAppId, loading: appsLoading } = useApp();
    const [userMenuOpen, setUserMenuOpen] = useState(false);
    const userMenuRef = useRef<HTMLDivElement>(null);

    // Close menu on outside click
    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (userMenuRef.current && !userMenuRef.current.contains(event.target as Node)) {
                setUserMenuOpen(false);
            }
        };

        if (userMenuOpen) {
            document.addEventListener('mousedown', handleClickOutside);
        }

        return () => {
            document.removeEventListener('mousedown', handleClickOutside);
        };
    }, [userMenuOpen]);

    const handleLogout = () => {
        logout();
        setUserMenuOpen(false);
    };

    const getUserInitials = () => {
        if (!user?.name) return 'U';
        const parts = user.name.split(' ');
        if (parts.length >= 2) {
            return `${parts[0][0]}${parts[1][0]}`.toUpperCase();
        }
        return user.name[0].toUpperCase();
    };

    return (
        <header className="h-16 bg-white border-b border-slate-200 shadow-sm z-50 flex items-center justify-between px-4 shrink-0">
            <div className="flex items-center gap-4">
                <div className="flex items-center gap-3">
                    {/* Mobile Menu Toggle */}
                    <button
                        onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
                        className="lg:hidden p-2 -ml-2 text-slate-500 hover:bg-slate-100 rounded-lg transition-colors"
                    >
                        <Menu size={24} />
                    </button>

                    {/* Logo */}
                    <div className={`font-bold text-xl text-slate-800 tracking-tight flex items-center gap-2 ${isSidebarCollapsed ? '' : 'lg:w-52'}`}>
                        <div className="w-8 h-8 rounded-lg flex items-center justify-center text-white" style={{ backgroundColor: 'var(--color-brand)' }}>
                            <span className="font-bold text-lg">N</span>
                        </div>
                        <span className="inline">NexusCRM</span>
                    </div>
                </div>

                {/* App Launcher */}
                {!appsLoading && (
                    <AppLauncher
                        currentAppId={currentAppId}
                        onSelectApp={(appId) => {
                            localStorage.setItem(STORAGE_KEYS.CURRENT_APP, appId);
                            window.location.href = '/';
                        }}
                    />
                )}
            </div>

            <div className="flex items-center gap-3 flex-1 max-w-2xl mx-4">
                <GlobalSearch />
            </div>

            <div className="flex items-center gap-4">
                {/* AI Assistant Toggle */}
                <button
                    onClick={onToggleAI}
                    className="group flex items-center gap-2 px-3 py-2 bg-white hover:bg-slate-50 text-slate-700 rounded-full transition-all border border-slate-200 shadow-sm"
                    title="Nexus AI"
                >
                    <Bot size={20} className="group-hover:text-indigo-600 transition-colors" />
                    <span className="font-semibold text-sm">Nexus AI</span>
                </button>

                <NotificationCenter />

                <div className="relative" ref={userMenuRef}>
                    <button
                        onClick={() => setUserMenuOpen(!userMenuOpen)}
                        className="w-9 h-9 bg-slate-200 rounded-full flex items-center justify-center text-slate-600 font-medium hover:ring-2 hover:ring-offset-2 hover:ring-blue-500 transition-all"
                    >
                        {getUserInitials()}
                    </button>
                    {/* Dropdown Menu */}
                    {userMenuOpen && (
                        <div className="absolute right-0 mt-2 w-56 bg-white border border-slate-200 rounded-lg shadow-lg py-1 z-20">
                            <div className="px-4 py-3 border-b border-slate-100">
                                <p className="text-sm font-semibold text-slate-900">{user?.name || 'User'}</p>
                                {user?.email && (
                                    <p className="text-xs text-slate-500 mt-0.5">{user.email}</p>
                                )}
                            </div>

                            <button
                                onClick={handleLogout}
                                className="w-full px-4 py-2 text-left text-sm text-slate-700 hover:bg-slate-50 flex items-center gap-2 transition-colors"
                            >
                                <LogOut size={16} />
                                <span>Logout</span>
                            </button>
                        </div>
                    )}
                </div>
            </div>
        </header>
    );
};
