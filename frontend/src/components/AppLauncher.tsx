import React, { useState, useRef, useEffect } from 'react';
import { Link } from 'react-router-dom';
import * as Icons from 'lucide-react';
import { Grid3X3, ChevronDown, Check } from 'lucide-react';
import { useApp } from '../contexts/AppContext';

interface AppLauncherProps {
    currentAppId: string | null;
    onSelectApp: (appId: string) => void;
}

export const AppLauncher: React.FC<AppLauncherProps> = ({ currentAppId, onSelectApp }) => {
    const [isOpen, setIsOpen] = useState(false);
    const { visibleApps, loading } = useApp();
    const dropdownRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        const handleClickOutside = (event: MouseEvent) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
                setIsOpen(false);
            }
        };

        if (isOpen) {
            document.addEventListener('mousedown', handleClickOutside);
        }
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, [isOpen]);

    const currentApp = visibleApps.find(app => app.id === currentAppId);

    const getIconComponent = (iconName: string) => {
        const IconComponent = Icons[iconName as keyof typeof Icons] as React.ComponentType<any>;
        return IconComponent || Icons.Layers;
    };

    return (
        <div className="relative" ref={dropdownRef}>
            <button
                onClick={() => setIsOpen(!isOpen)}
                className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-slate-100 transition-colors"
                title="App Launcher"
            >
                <Grid3X3 size={20} className="text-slate-600" />
                {currentApp && (
                    <>
                        <span className="text-sm font-medium text-slate-700 hidden md:block">
                            {currentApp.label}
                        </span>
                        <ChevronDown size={16} className={`text-slate-500 transition-transform ${isOpen ? 'rotate-180' : ''}`} />
                    </>
                )}
            </button>

            {isOpen && (
                <div className="absolute left-0 mt-2 w-80 bg-white border border-slate-200 rounded-lg shadow-xl z-50">
                    <div className="p-3 border-b border-slate-100">
                        <h3 className="text-sm font-semibold text-slate-700">App Launcher</h3>
                        <p className="text-xs text-slate-500 mt-0.5">Switch between applications</p>
                    </div>

                    <div className="p-2 max-h-80 overflow-y-auto">
                        {loading ? (
                            <div className="p-4 text-center text-slate-500 text-sm">Loading apps...</div>
                        ) : visibleApps.length === 0 ? (
                            <div className="p-3 space-y-3">
                                <p className="text-xs text-slate-500 text-center">No apps created yet</p>

                                {/* Quick Links */}
                                <div className="space-y-1">
                                    <p className="text-xs font-semibold text-slate-600 uppercase tracking-wide">Quick Links</p>
                                    <Link
                                        to="/setup"
                                        onClick={() => setIsOpen(false)}
                                        className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-slate-50 transition-colors"
                                    >
                                        <Icons.Settings size={16} className="text-slate-500" />
                                        <span className="text-sm text-slate-700">Setup</span>
                                    </Link>
                                    <Link
                                        to="/setup/objects"
                                        onClick={() => setIsOpen(false)}
                                        className="flex items-center gap-2 px-3 py-2 rounded-lg hover:bg-slate-50 transition-colors"
                                    >
                                        <Icons.Database size={16} className="text-slate-500" />
                                        <span className="text-sm text-slate-700">Object Manager</span>
                                    </Link>
                                </div>

                                {/* Create App CTA */}
                                <Link
                                    to="/setup/appmanager"
                                    onClick={() => setIsOpen(false)}
                                    className="flex items-center justify-center gap-2 w-full px-3 py-2 bg-blue-50 text-blue-600 rounded-lg hover:bg-blue-100 transition-colors"
                                >
                                    <Icons.Plus size={16} />
                                    <span className="text-sm font-medium">Create Your First App</span>
                                </Link>
                            </div>
                        ) : (
                            <div className="grid grid-cols-3 gap-2">
                                {visibleApps.map(app => {
                                    const IconComponent = getIconComponent(app.icon);
                                    const isSelected = app.id === currentAppId;
                                    return (
                                        <button
                                            key={app.id}
                                            onClick={() => {
                                                onSelectApp(app.id);
                                                setIsOpen(false);
                                            }}
                                            className={`flex flex-col items-center p-3 rounded-lg transition-colors relative
                        ${isSelected
                                                    ? 'bg-blue-50 ring-2 ring-blue-500'
                                                    : 'hover:bg-slate-50'
                                                }`}
                                        >
                                            {isSelected && (
                                                <div className="absolute top-1 right-1">
                                                    <Check size={12} className="text-blue-600" />
                                                </div>
                                            )}
                                            <div className={`w-10 h-10 rounded-lg flex items-center justify-center mb-1.5 shrink-0
                        ${isSelected ? 'bg-blue-100 text-blue-600' : 'bg-slate-100 text-slate-600'}`}
                                            >
                                                <IconComponent size={20} />
                                            </div>
                                            <span className={`text-xs font-medium text-center line-clamp-2 break-words w-full leading-tight
                        ${isSelected ? 'text-blue-700' : 'text-slate-700'}`}>
                                                {app.label}
                                            </span>
                                        </button>
                                    );
                                })}
                            </div>
                        )}
                    </div>

                    <div className="p-2 border-t border-slate-100">
                        <Link
                            to="/setup/appmanager"
                            onClick={() => setIsOpen(false)}
                            className="block w-full text-center text-sm text-blue-600 hover:text-blue-700 py-2 hover:bg-blue-50 rounded-lg transition-colors"
                        >
                            Manage Apps
                        </Link>
                    </div>
                </div>
            )}
        </div>
    );
};
