import React, { useState, useEffect } from 'react';
import { useErrorToast } from '../ui/Toast';
import { metadataAPI } from '../../infrastructure/api/metadata';
import { DashboardConfig, WidgetConfig, ObjectMetadata } from '../../types';
import { WIDGET_SIZE_DEFAULTS, DEFAULT_WIDGET_SIZE } from '../../core/constants/widgets';
import { DashboardPalette } from './dashboard/DashboardPalette';
import { DashboardInspector } from './dashboard/DashboardInspector';
import { DashboardWidgetSkeleton } from '../ui/LoadingSkeleton';
import { Layout } from 'react-grid-layout';

// Sub-components
import { DashboardToolbar } from './dashboard/DashboardToolbar';
import { DashboardCanvas } from './dashboard/DashboardCanvas';

interface StudioDashboardEditorProps {
    dashboardId: string;
    onDashboardUpdated?: () => void;
}

export const StudioDashboardEditor: React.FC<StudioDashboardEditorProps> = ({ dashboardId, onDashboardUpdated }) => {
    const errorToast = useErrorToast();
    const [title, setTitle] = useState('');
    const [widgets, setWidgets] = useState<WidgetConfig[]>([]);
    const [schemas, setSchemas] = useState<ObjectMetadata[]>([]);

    // UI State
    const [saving, setSaving] = useState(false);
    const [loading, setLoading] = useState(true);
    const [selectedWidgetId, setSelectedWidgetId] = useState<string | null>(null);
    const [isDropping, setIsDropping] = useState(false);

    useEffect(() => {
        loadData();
    }, [dashboardId]);

    const loadData = async () => {
        setLoading(true);
        try {
            const schemasRes = await metadataAPI.getSchemas();
            setSchemas(schemasRes.schemas);

            const res = await metadataAPI.getDashboard(dashboardId);
            if (res.dashboard) {
                setTitle(res.dashboard.label);
                setWidgets(res.dashboard.widgets || []);
            }
        } catch {
            // Dashboard loading failure handled via UI empty state
        } finally {
            setLoading(false);
        }
    };

    const handleSaveDashboard = async () => {
        if (!title) return errorToast('Please enter a dashboard title');
        setSaving(true);
        try {
            const dashboardPayload: DashboardConfig = {
                id: dashboardId,
                label: title,
                widgets: widgets
            };
            await metadataAPI.updateDashboard(dashboardId, dashboardPayload);
            if (onDashboardUpdated) onDashboardUpdated();
        } catch {
            errorToast('Failed to save dashboard');
        } finally {
            setSaving(false);
        }
    };

    // --- Grid Layout Handlers ---

    const handleLayoutChange = (layout: Layout[]) => {
        // Only update if not currently dropping, to avoid race conditions
        if (isDropping) return;

        const updatedWidgets = widgets.map(w => {
            const l = layout.find(item => item.i === w.id);
            if (l) {
                return { ...w, x: l.x, y: l.y, w: l.w, h: l.h };
            }
            return w;
        });

        // Deep verify change to prevent cycles
        const hasChanged = JSON.stringify(updatedWidgets.map(w => ({ x: w.x, y: w.y, w: w.w, h: w.h }))) !==
            JSON.stringify(widgets.map(w => ({ x: w.x, y: w.y, w: w.w, h: w.h })));

        if (hasChanged) {
            setWidgets(updatedWidgets);
        }
    };

    const handleDrop = (layout: Layout[], layoutItem: Layout, _event: DragEvent) => {
        setIsDropping(true);
        const widgetType = _event.dataTransfer?.getData('text/plain') || 'metric';

        const newId = `w_${Date.now()}`;
        const defaultTitle = widgetType === 'text' ? 'Text Block' : 'New Widget';

        const defaultW = WIDGET_SIZE_DEFAULTS[widgetType]?.w || DEFAULT_WIDGET_SIZE.w;
        const defaultH = WIDGET_SIZE_DEFAULTS[widgetType]?.h || DEFAULT_WIDGET_SIZE.h;

        const newWidget: WidgetConfig = {
            id: newId,
            title: defaultTitle,
            type: widgetType,
            query: { object_api_name: schemas[0]?.api_name || '', operation: 'count' },
            x: layoutItem.x,
            y: layoutItem.y,
            w: defaultW,
            h: defaultH,
            config: {
                content: widgetType === 'text' ? '### New Text Block\nDouble click inspector to edit.' : undefined,
                imageUrl: widgetType === 'image' ? 'https://via.placeholder.com/300' : undefined
            }
        };

        setWidgets(prev => [...prev, newWidget]);
        setSelectedWidgetId(newId);
        setIsDropping(false);
    };

    const handleAddWidget = (widgetType: string = 'metric') => {
        // Since we are adding by click, we don't have a drop position.
        // We'll calculate the lowest Y position to append to bottom.
        let maxY = 0;
        widgets.forEach(w => {
            const bottom = (w.y || 0) + (w.h || 1);
            if (bottom > maxY) maxY = bottom;
        });

        const newId = `w_${Date.now()}`;
        const defaultTitle = widgetType === 'text' ? 'New Text Block' :
            widgetType === 'image' ? 'New Image' :
                'New Widget';

        const defaultW = WIDGET_SIZE_DEFAULTS[widgetType]?.w || DEFAULT_WIDGET_SIZE.w;
        const defaultH = WIDGET_SIZE_DEFAULTS[widgetType]?.h || DEFAULT_WIDGET_SIZE.h;

        const newWidget: WidgetConfig = {
            id: newId,
            title: defaultTitle,
            type: widgetType,
            query: { object_api_name: schemas[0]?.api_name || '', operation: 'count' },
            x: 0,
            y: maxY, // Put at bottom
            w: defaultW,
            h: defaultH,
            config: {
                content: widgetType === 'text' ? '### New Text Block\nDouble click inspector to edit.' : undefined,
                imageUrl: widgetType === 'image' ? 'https://via.placeholder.com/300' : undefined
            }
        };
        setWidgets(prev => [...prev, newWidget]);
        setSelectedWidgetId(newId);
    };

    const handleWidgetUpdate = (id: string, updates: Partial<WidgetConfig>) => {
        setWidgets(prev => prev.map(w => w.id === id ? { ...w, ...updates } : w));
    };

    const onInspectorUpdate = (updates: Partial<WidgetConfig>) => {
        if (!selectedWidgetId) return;
        handleWidgetUpdate(selectedWidgetId, updates);
    };

    const handleWidgetDelete = () => {
        if (!selectedWidgetId) return;
        setWidgets(widgets.filter(w => w.id !== selectedWidgetId));
        setSelectedWidgetId(null);
    };

    if (loading) return (
        <div className="h-full flex items-center justify-center bg-slate-100 p-8">
            <div className="grid grid-cols-3 gap-6 w-full max-w-4xl">
                <DashboardWidgetSkeleton />
                <DashboardWidgetSkeleton />
                <DashboardWidgetSkeleton />
            </div>
        </div>
    );

    const selectedWidget = widgets.find(w => w.id === selectedWidgetId) || null;

    return (
        <div className="flex h-full bg-slate-100 overflow-hidden">
            {/* Left Palette */}
            <DashboardPalette onAddWidget={handleAddWidget} />

            {/* Center Canvas */}
            <div className="flex-1 flex flex-col min-w-0">
                <DashboardToolbar
                    title={title}
                    setTitle={setTitle}
                    onAddWidget={handleAddWidget}
                    onDeleteWidget={handleWidgetDelete}
                    onSave={handleSaveDashboard}
                    saving={saving}
                    hasSelectedWidget={!!selectedWidget}
                />

                <DashboardCanvas
                    widgets={widgets}
                    selectedWidgetId={selectedWidgetId}
                    onSelectWidget={setSelectedWidgetId}
                    onLayoutChange={handleLayoutChange}
                    onDrop={handleDrop}
                    onWidgetUpdate={handleWidgetUpdate}
                />
            </div>

            {/* Right Inspector */}
            <DashboardInspector
                selectedWidget={selectedWidget}
                schemas={schemas}
                onUpdate={onInspectorUpdate}
            />
        </div>
    );
};
