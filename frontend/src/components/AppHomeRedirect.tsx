import React, { useEffect, useState } from 'react';
import { Navigate } from 'react-router-dom';
import { useApp } from '../contexts/AppContext';
import { Dashboard } from '../pages/Dashboard';

export const AppHomeRedirect: React.FC = () => {
    const { currentApp, loading } = useApp();
    const [shouldRedirect, setShouldRedirect] = useState(false);
    const [redirectPath, setRedirectPath] = useState<string | null>(null);

    useEffect(() => {
        if (!loading && currentApp) {
            // Find first valid navigation item
            if (currentApp.navigation_items && currentApp.navigation_items.length > 0) {
                const firstItem = currentApp.navigation_items[0];

                // Construct path based on item type
                // Usually: /object/:apiName or /page/:pageId
                let path = null;
                if (firstItem.type === 'object' && firstItem.object_api_name) {
                    path = `/object/${firstItem.object_api_name}`;
                } else if (firstItem.type === 'web' && firstItem.page_url) {
                    // Assuming internal pages might use this or handle external
                    path = firstItem.page_url.startsWith('/') ? firstItem.page_url : `/web/${encodeURIComponent(firstItem.page_url)}`;
                } else if (firstItem.type === 'dashboard' && firstItem.dashboard_id) {
                    path = `/dashboard/${firstItem.dashboard_id}`;
                }

                if (path) {
                    setRedirectPath(path);
                    setShouldRedirect(true);
                }
            }
        }
    }, [currentApp, loading]);

    if (loading) {
        return <div className="p-8 text-center text-slate-500">Loading app...</div>;
    }

    if (shouldRedirect && redirectPath) {
        return <Navigate to={redirectPath} replace />;
    }

    // Fallback if no navigation items or no app selected
    return <Dashboard />;
};
