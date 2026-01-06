import React, { useMemo, useCallback } from 'react';
import { Routes, Route, Link, Navigate, useLocation } from 'react-router-dom';
import { SetupRegistry, SetupPageDefinition } from '../registries/SetupRegistry';
import { ComponentRegistry } from '../registries/ComponentRegistry';
import { Settings as SettingsIcon, AlertCircle } from 'lucide-react';
import * as LucideIcons from 'lucide-react';

// ============================================================================
// Shared Hook - Single source of truth for loading Setup pages
// ============================================================================

function useSetupPages() {
    const [pages, setPages] = React.useState<SetupPageDefinition[]>(
        SetupRegistry.isLoaded() ? SetupRegistry.getPages() : []
    );
    const [loading, setLoading] = React.useState(!SetupRegistry.isLoaded());

    React.useEffect(() => {
        if (SetupRegistry.isLoaded()) {
            setPages(SetupRegistry.getPages());
            return;
        }

        const load = async () => {
            await SetupRegistry.loadFromDatabase();
            setPages(SetupRegistry.getPages());
            setLoading(false);
        };
        load();
    }, []);

    return { pages, loading };
}

// ============================================================================
// Category Colors - Could be metadata-driven in future
// ============================================================================

const CATEGORY_COLORS: Record<string, { bg: string; text: string; hover: string }> = {
    'Security': { bg: 'bg-emerald-100', text: 'text-emerald-600', hover: 'group-hover:bg-emerald-600' },
    'Automation': { bg: 'bg-purple-100', text: 'text-purple-600', hover: 'group-hover:bg-purple-600' },
    'System': { bg: 'bg-cyan-100', text: 'text-cyan-600', hover: 'group-hover:bg-cyan-600' },
    'Objects': { bg: 'bg-amber-100', text: 'text-amber-600', hover: 'group-hover:bg-amber-600' },
};
const DEFAULT_COLORS = { bg: 'bg-blue-100', text: 'text-blue-600', hover: 'group-hover:bg-blue-600' };

// ============================================================================
// Not Implemented Placeholder Component
// ============================================================================

const NotImplemented: React.FC<{ componentName: string; pageLabel: string }> = ({ componentName, pageLabel }) => (
    <div className="flex flex-col items-center justify-center p-12 bg-slate-50 rounded-xl border-2 border-dashed border-slate-300">
        <AlertCircle className="w-16 h-16 text-slate-400 mb-4" />
        <h2 className="text-xl font-semibold text-slate-700 mb-2">{pageLabel}</h2>
        <p className="text-slate-500 text-center max-w-md">
            This feature is coming soon. The <code className="bg-slate-200 px-2 py-1 rounded text-sm">{componentName}</code> component is registered in metadata but not yet implemented.
        </p>
    </div>
);

// ============================================================================
// Setup Home - Displays tiles from SetupRegistry
// ============================================================================

const SetupHome: React.FC = () => {
    const { pages, loading } = useSetupPages();

    // Group pages by category
    const categories = useMemo(() => {
        return pages.reduce((acc, page) => {
            const cat = page.category || 'Other';
            if (!acc[cat]) acc[cat] = [];
            acc[cat].push(page);
            return acc;
        }, {} as Record<string, SetupPageDefinition[]>);
    }, [pages]);

    if (loading) {
        return (
            <div className="flex items-center justify-center p-12">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    return (
        <div className="space-y-8">
            {Object.entries(categories).map(([category, categoryPages]) => (
                <div key={category}>
                    <h2 className="text-lg font-semibold text-slate-700 mb-4">{category}</h2>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {categoryPages.map((page) => {
                            const IconComponent = (LucideIcons as unknown as Record<string, React.FC<{ size?: number }>>)[page.icon] || LucideIcons.Box;
                            const colors = CATEGORY_COLORS[page.category || ''] || DEFAULT_COLORS;
                            const isImplemented = ComponentRegistry.has(page.component_name);

                            return (
                                <Link
                                    key={page.id}
                                    to={page.path || '#'}
                                    className={`p-6 bg-white/80 backdrop-blur-xl rounded-2xl border border-white/20 hover:border-blue-500 hover:shadow-xl transition-all group ${!isImplemented ? 'opacity-75' : ''}`}
                                >
                                    <div className={`w-10 h-10 ${colors.bg} rounded-lg flex items-center justify-center ${colors.text} mb-4 ${colors.hover} group-hover:text-white transition-colors`}>
                                        <IconComponent size={20} />
                                    </div>
                                    <div className="flex items-center gap-2 mb-2">
                                        <h3 className="text-lg font-semibold text-slate-800">{page.label}</h3>
                                        {!isImplemented && (
                                            <span className="text-xs bg-slate-200 text-slate-600 px-2 py-0.5 rounded">Coming Soon</span>
                                        )}
                                    </div>
                                    <p className="text-slate-500 text-sm">{page.description}</p>
                                </Link>
                            );
                        })}
                    </div>
                </div>
            ))}
        </div>
    );
};

// ============================================================================
// Dynamic Route Generator - Creates routes from metadata
// Uses ComponentRegistry as single source of truth for component resolution
// ============================================================================

const DynamicSetupRoutes: React.FC = () => {
    const { pages, loading } = useSetupPages();

    if (loading) {
        return (
            <div className="flex items-center justify-center p-12">
                <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            </div>
        );
    }

    return (
        <Routes>
            <Route path="/" element={<SetupHome />} />

            {/* Dynamic routes from metadata */}
            {pages.map((page) => {
                // Extract path relative to /setup (e.g., "/setup/users" -> "/users")
                const relativePath = page.path?.replace('/setup', '') || `/${page.component_name.toLowerCase()}`;
                const Component = ComponentRegistry.get(page.component_name);

                return (
                    <Route
                        key={page.id}
                        path={relativePath}
                        element={
                            Component
                                ? <Component />
                                : <NotImplemented componentName={page.component_name} pageLabel={page.label} />
                        }
                    />
                );
            })}

            {/* Static routes for nested pages (these are not top-level setup pages) */}
            <Route path="/objects/:objectApiName" element={<ObjectDetailRoute />} />
            <Route path="/objects/:objectApiName/layout" element={<LayoutEditorRoute />} />

            {/* URL Alias Redirects - Friendly URLs that redirect to actual component paths */}
            <Route path="/users" element={<Navigate to="/setup/usermanager" replace />} />
            <Route path="/profiles" element={<Navigate to="/setup/usermanager" replace />} />
            <Route path="/sharing-rules" element={<Navigate to="/setup/sharingrulemanager" replace />} />

            {/* Explicit Object Manager Route to prevent 404s (BUG-007 fix) */}
            <Route path="/objects" element={<ObjectManagerRoute />} />

            {/* Catch-all for unknown routes */}
            <Route path="*" element={
                <div className="text-center p-12 text-slate-500">
                    <AlertCircle className="w-12 h-12 mx-auto mb-4 text-slate-400" />
                    <p>Page not found in Setup. <Link to="/setup" className="text-blue-600 hover:underline">Go back to Setup</Link></p>
                </div>
            } />
        </Routes>
    );
};

// Wrapper components for nested routes to use ComponentRegistry
const ObjectDetailRoute: React.FC = () => {
    const Component = ComponentRegistry.get('ObjectDetail');
    return Component ? <Component /> : null;
};

const LayoutEditorRoute: React.FC = () => {
    const Component = ComponentRegistry.get('LayoutEditor');
    return Component ? <Component /> : null;
};

const ObjectManagerRoute: React.FC = () => {
    const Component = ComponentRegistry.get('ObjectManager');
    return Component ? <Component /> : <NotImplemented componentName="ObjectManager" pageLabel="Object Manager" />;
};

// ============================================================================
// Main Setup Component
// ============================================================================

export const Setup: React.FC = () => {
    const location = useLocation();
    const isHome = location.pathname === '/setup';

    // Memoized page label lookup
    const currentPageLabel = useMemo(() => {
        if (isHome) return 'Setup & Configuration';

        const pages = SetupRegistry.getPages();
        const currentPage = pages.find(p => p.path === location.pathname);
        if (currentPage) return currentPage.label;

        // Fallback: capitalize last segment
        const segment = location.pathname.split('/').pop() || '';
        return segment.charAt(0).toUpperCase() + segment.slice(1);
    }, [location.pathname, isHome]);

    return (
        <div className="max-w-7xl mx-auto">
            <div className="flex items-center gap-2 mb-6 text-sm text-slate-500">
                <Link to="/setup" className="hover:text-blue-600 flex items-center gap-1">
                    <SettingsIcon size={14} /> Setup
                </Link>
                {!isHome && (
                    <>
                        <span>/</span>
                        <span className="text-slate-800 font-medium">
                            {currentPageLabel}
                        </span>
                    </>
                )}
            </div>

            <h1 className="text-2xl font-bold text-slate-800 mb-6">
                {currentPageLabel}
            </h1>

            <DynamicSetupRoutes />
        </div>
    );
};
