import React from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { DashboardWidgetRegistry } from '../registries/DashboardWidgetRegistry';
import { metadataAPI } from '../infrastructure/api/metadata';
import { useApp } from '../contexts/AppContext';
import { LayoutDashboard, TrendingUp, Settings, Plus, Edit, Trash2, ArrowLeft, RefreshCw, Clock } from 'lucide-react';
import type { DashboardConfig, WidgetConfig } from '../types';
import { WIDGET_SIZE_DEFAULTS, DEFAULT_WIDGET_SIZE } from '../core/constants/widgets';
import { ROUTES, BREAKPOINTS, GRID_COLS } from '../core/constants';
import { FilterBar, GlobalFilters } from '../components/FilterBar';
import { Responsive, WidthProvider } from 'react-grid-layout';
import 'react-grid-layout/css/styles.css';
import { ConfirmationModal } from '../components/modals/ConfirmationModal';
import { useErrorToast } from '../components/ui/Toast';
import { DashboardWidgetSkeleton } from '../components/ui/LoadingSkeleton';

const ResponsiveGridLayout = WidthProvider(Responsive);

/**
 * Dashboard Library - Hub for viewing and managing dashboards
 */
export const Dashboard: React.FC = () => {
  const errorToast = useErrorToast();
  const navigate = useNavigate();
  const { dashboardId } = useParams();
  const { currentAppId } = useApp();
  const [dashboards, setDashboards] = React.useState<DashboardConfig[]>([]);
  const [currentDashboard, setCurrentDashboard] = React.useState<DashboardConfig | null>(null);
  const [loading, setLoading] = React.useState(true);

  const [globalFilters, setGlobalFilters] = React.useState<GlobalFilters>({});
  const [dashboardToDelete, setDashboardToDelete] = React.useState<string | null>(null);
  const [showAppContextModal, setShowAppContextModal] = React.useState(false);
  const [lastRefresh, setLastRefresh] = React.useState<Date>(new Date());
  const [refreshing, setRefreshing] = React.useState(false);

  const handleRefresh = () => {
    setRefreshing(true);
    setLastRefresh(new Date());
    // Force re-render of widgets by updating timestamp
    setTimeout(() => setRefreshing(false), 500);
  };

  const fetchDashboards = () => {
    setLoading(true);
    metadataAPI.getDashboards()
      .then(res => {
        setDashboards(res.dashboards || []);
        setLoading(false);
      })
      .catch(() => {
        // Dashboard loading failure - handled via empty state
        setLoading(false);
      });
  };

  React.useEffect(() => {
    fetchDashboards();
  }, []);

  React.useEffect(() => {
    if (dashboards.length > 0 && dashboardId) {
      const found = dashboards.find(d => d.id === dashboardId);
      if (found) setCurrentDashboard(found);
    } else if (!dashboardId) {
      setCurrentDashboard(null);
    }
  }, [dashboardId, dashboards]);

  const handleDelete = (e: React.MouseEvent, id: string) => {
    e.stopPropagation();
    setDashboardToDelete(id);
  };

  const executeDelete = async () => {
    if (!dashboardToDelete) return;
    try {
      await metadataAPI.deleteDashboard(dashboardToDelete);
      setDashboards(prev => prev.filter(d => d.id !== dashboardToDelete));
      if (currentDashboard?.id === dashboardToDelete) setCurrentDashboard(null);
    } catch (err) {
      errorToast('Failed to delete dashboard');
    } finally {
      setDashboardToDelete(null);
    }
  };

  const handleEdit = (e: React.MouseEvent, id: string) => {
    e.stopPropagation();
    if (currentAppId) {
      navigate(`/studio/${currentAppId}?dashboardId=${id}`);
    } else {
      navigate(ROUTES.SETUP.APPS);
    }
  };

  if (loading && !dashboards.length) return (
    <div className="p-8 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <DashboardWidgetSkeleton />
      <DashboardWidgetSkeleton />
      <DashboardWidgetSkeleton />
    </div>
  );

  return (
    <div className="flex-1 overflow-auto bg-slate-50/50 min-h-screen flex flex-col">
      {/* Header */}
      <div className="sticky top-0 z-20 bg-white/80 backdrop-blur-md border-b border-slate-200 px-8 py-4 flex items-center justify-between shadow-sm">
        <div className="flex items-center gap-4">
          {currentDashboard ? (
            <>
              <button
                onClick={() => navigate('/dashboards')}
                className="p-2 hover:bg-slate-100 rounded-lg text-slate-500 hover:text-slate-700 transition-colors"
                title="Back to Dashboards"
              >
                <ArrowLeft size={20} />
              </button>
              <div>
                <h1 className="text-xl font-bold text-slate-800">{currentDashboard.label}</h1>
                <p className="text-sm text-slate-500 hidden sm:block">Viewing Dashboard</p>
              </div>
            </>
          ) : (
            <div>
              <h1 className="text-xl font-bold text-slate-800">My Dashboards</h1>
              <p className="text-sm text-slate-500 hidden sm:block">Manage and view your analytics</p>
            </div>
          )}
        </div>

        <div className="flex gap-3">
          {!currentDashboard && (
            <button
              onClick={() => {
                if (currentAppId) {
                  navigate(`/studio/${currentAppId}?action=add_page`);
                } else {
                  setShowAppContextModal(true);
                }
              }}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg shadow-sm hover:bg-blue-700 hover:shadow font-medium flex items-center gap-2 transition-all"
            >
              <Plus size={18} />
              New Dashboard
            </button>
          )}
          {currentDashboard && (
            <>
              <div className="flex items-center gap-2 text-sm text-slate-500 mr-2">
                <Clock size={14} />
                <span>Updated {lastRefresh.toLocaleTimeString()}</span>
              </div>
              <button
                onClick={handleRefresh}
                disabled={refreshing}
                className="p-2 bg-white border border-slate-200 text-slate-700 rounded-lg shadow-sm hover:bg-slate-50 transition-all disabled:opacity-50"
                title="Refresh Dashboard"
              >
                <RefreshCw size={16} className={refreshing ? 'animate-spin' : ''} />
              </button>
              <button
                onClick={(e) => handleEdit(e, currentDashboard.id)}
                className="px-4 py-2 bg-white border border-slate-200 text-slate-700 rounded-lg shadow-sm hover:bg-slate-50 font-medium flex items-center gap-2 transition-all"
              >
                <Edit size={16} />
                Edit
              </button>
            </>
          )}
        </div>
      </div>

      {/* Main Content */}
      <div className="p-8">
        {!currentDashboard ? (
          /* Hub View */
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {dashboards.length === 0 ? (
              <div className="col-span-full flex flex-col items-center justify-center py-24 bg-white border-2 border-dashed border-slate-200 rounded-2xl">
                <div className="w-16 h-16 bg-slate-50 rounded-full flex items-center justify-center mb-4 text-slate-400">
                  <LayoutDashboard size={32} strokeWidth={1.5} />
                </div>
                <h3 className="text-lg font-medium text-slate-900">No dashboards yet</h3>
                <p className="text-slate-500 mt-1 mb-6 max-w-sm text-center">Create your first dashboard to start visualizing your data.</p>
                <button
                  onClick={() => {
                    if (currentAppId) {
                      navigate(`/studio/${currentAppId}?tab=dashboards`);
                    } else {
                      setShowAppContextModal(true);
                    }
                  }}
                  className="px-5 py-2.5 bg-blue-600 text-white rounded-xl hover:bg-blue-700 font-medium flex items-center gap-2 transition-all shadow-lg shadow-blue-600/20"
                >
                  Create Dashboard
                </button>
              </div>
            ) : (
              dashboards.map(d => (
                <div
                  key={d.id}
                  onClick={() => navigate(`/dashboard/${d.id}`)}
                  className="group bg-white rounded-2xl border border-slate-200 shadow-sm hover:shadow-lg hover:-translate-y-1 transition-all duration-200 cursor-pointer flex flex-col h-56 relative overflow-hidden"
                >
                  <div className="h-2 bg-gradient-to-r from-blue-500 to-indigo-500 opacity-0 group-hover:opacity-100 transition-opacity absolute top-0 left-0 right-0"></div>

                  <div className="p-6 flex-1">
                    <div className="flex justify-between items-start mb-4">
                      <div className="p-3 bg-blue-50 text-blue-600 rounded-xl group-hover:scale-110 transition-transform duration-200">
                        <LayoutDashboard size={24} />
                      </div>
                    </div>
                    <h3 className="font-bold text-slate-800 text-lg group-hover:text-blue-600 transition-colors mb-1 truncate">{d.label}</h3>
                    <p className="text-sm text-slate-400 line-clamp-2">{d.description || "No description provided."}</p>
                  </div>

                  <div className="px-6 py-4 bg-slate-50 border-t border-slate-100 flex justify-between items-center">
                    <span className="text-xs font-medium text-slate-500 flex items-center gap-1.5">
                      <TrendingUp size={14} />
                      {d.widgets?.length || 0} Widgets
                    </span>
                    <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button
                        onClick={(e) => handleEdit(e, d.id)}
                        className="p-2 hover:bg-white rounded-lg text-slate-400 hover:text-blue-600 hover:shadow-sm transition-all"
                        title="Edit in Studio"
                      >
                        <Settings size={16} />
                      </button>
                      <button
                        onClick={(e) => handleDelete(e, d.id)}
                        className="p-2 hover:bg-white rounded-lg text-slate-400 hover:text-red-600 hover:shadow-sm transition-all"
                        title="Delete"
                      >
                        <Trash2 size={16} />
                      </button>
                    </div>
                  </div>
                </div>
              ))
            )}
          </div>
        ) : (
          /* Detail View */
          <div className="space-y-6">
            <FilterBar filters={globalFilters} onFilterChange={setGlobalFilters} />

            <ResponsiveGridLayout
              className="layout"
              layouts={{
                lg: currentDashboard.widgets?.map((w: WidgetConfig) => {
                  const defaults = WIDGET_SIZE_DEFAULTS[w.type] || DEFAULT_WIDGET_SIZE;

                  const width = w.w || defaults.w;
                  const height = w.h || defaults.h;

                  return {
                    i: w.id,
                    x: w.x || 0,
                    y: w.y || 0,
                    w: width,
                    h: height
                  };
                }) || []
              }}
              breakpoints={BREAKPOINTS}
              cols={GRID_COLS}
              rowHeight={80}
              isDraggable={false}
              isResizable={false}
              margin={[20, 20]}
            >
              {currentDashboard.widgets?.map((widget: WidgetConfig) => {
                const props = {
                  ...widget,
                  id: widget.id,
                  title: widget.title,
                  data: [],
                  loading: false,
                  config: widget,
                  globalFilters,
                  isEditing: false
                };
                const WidgetComponent = DashboardWidgetRegistry.getWidget(widget.type);

                return (
                  <div key={widget.id} className="bg-white rounded-xl shadow-sm border border-slate-200 overflow-hidden h-full hover:shadow-lg transition-all duration-200">
                    {WidgetComponent ? (
                      <WidgetComponent {...props} />
                    ) : (
                      <div className="flex items-center justify-center h-full text-slate-400 text-sm">
                        <span className="p-4 text-center">Unknown Widget: {widget.type}</span>
                      </div>
                    )}
                  </div>
                );
              })}
            </ResponsiveGridLayout>

          </div>
        )}
      </div>
      <ConfirmationModal
        isOpen={!!dashboardToDelete}
        onClose={() => setDashboardToDelete(null)}
        onConfirm={executeDelete}
        title="Delete Dashboard"
        message="Are you sure you want to delete this dashboard? This action cannot be undone."
        confirmLabel="Delete"
        variant="danger"
      />
      <ConfirmationModal
        isOpen={showAppContextModal}
        onClose={() => setShowAppContextModal(false)}
        onConfirm={() => {
          setShowAppContextModal(false);
          navigate(ROUTES.SETUP.APPS);
        }}
        title="App Context Required"
        message="You need to be in an App context to create a dashboard. Would you like to go to App Manager?"
        confirmLabel="Go to App Manager"
        variant="info"
      />
    </div >
  );
};
