import React from 'react';
import { Responsive, WidthProvider, Layout } from 'react-grid-layout';
import 'react-grid-layout/css/styles.css';
import 'react-resizable/css/styles.css';
import ReactMarkdown from 'react-markdown';
import { DashboardWidgetRegistry } from '../../../registries/DashboardWidgetRegistry';
import { WidgetConfig } from '../../../types';

const ResponsiveGridLayout = WidthProvider(Responsive);

interface DashboardCanvasProps {
    widgets: WidgetConfig[];
    selectedWidgetId: string | null;
    onSelectWidget: (id: string) => void;
    onLayoutChange: (layout: Layout[]) => void;
    onDrop: (layout: Layout[], layoutItem: Layout, event: DragEvent) => void;
    onWidgetUpdate: (id: string, updates: Partial<WidgetConfig>) => void;
}

export const DashboardCanvas: React.FC<DashboardCanvasProps> = ({
    widgets,
    selectedWidgetId,
    onSelectWidget,
    onLayoutChange,
    onDrop,
    onWidgetUpdate,
}) => {

    const renderWidgetContent = (widget: WidgetConfig) => {

        // Special Non-Data Widgets
        if (widget.type === 'text') {
            const content = typeof widget.config?.content === 'string' ? widget.config.content : '';
            return (
                <div className="p-4 h-full overflow-auto prose prose-sm max-w-none">
                    <ReactMarkdown>{content}</ReactMarkdown>
                </div>
            );
        }
        if (widget.type === 'image') {
            const imageUrl = typeof widget.config?.imageUrl === 'string' ? widget.config.imageUrl : '';
            return (
                <div className="h-full w-full flex items-center justify-center overflow-hidden bg-slate-50">
                    {imageUrl ? (
                        <img src={imageUrl} alt={widget.title} className="max-w-full max-h-full object-contain" />
                    ) : (
                        <div className="text-slate-400">No Image URL</div>
                    )}
                </div>
            );
        }

        // Data Widgets
        const commonProps = {
            ...widget,
            data: [], // In editor, we might show mock data or real data. For simplicity: mock/empty unless we fetch.
            loading: false,
            config: widget,

            isEditing: true,
            onConfigUpdate: (updates: Partial<WidgetConfig>) => {
                onWidgetUpdate(widget.id, updates);
            }
        };

        const WidgetComponent = DashboardWidgetRegistry.getWidget(widget.type);
        if (WidgetComponent) {
            return <WidgetComponent {...commonProps} />;
        }
        return <div className="p-4 text-red-500">Unknown Widget: {widget.type}</div>;
    };

    return (
        <div
            className="flex-1 overflow-auto p-8 relative"
            onDragOver={(e) => {
                e.preventDefault();
            }}
            onDrop={(e) => {
                // Handled by RGL onDrop usually, but we keep this for native drops if needed
            }}
        >
            <ResponsiveGridLayout
                className="layout"
                style={{ minHeight: '800px' }}
                layouts={{ lg: widgets.map(w => ({ i: w.id, x: w.x || 0, y: w.y || 0, w: w.w || 1, h: w.h || 1 })) }}
                breakpoints={{ lg: 1200, md: 996, sm: 768, xs: 480, xxs: 0 }}
                cols={{ lg: 12, md: 10, sm: 6, xs: 4, xxs: 2 }}
                rowHeight={60}
                onLayoutChange={onLayoutChange}
                onDrop={(layout, item, e) => {
                    // RGL types define e as Event, but it's a DragEvent in practice
                    onDrop(layout, item, e as unknown as DragEvent);
                }}
                isDroppable={true}
                resizeHandles={['se']}
                droppingItem={{ i: 'dropping', w: 2, h: 2 }}
                draggableHandle=".drag-handle"
            >
                {widgets.map(widget => (
                    <div
                        key={widget.id}
                        className={`
                            relative bg-white rounded-lg shadow-sm border transaction-all group hover:shadow-md
                            ${selectedWidgetId === widget.id ? 'border-blue-500 ring-2 ring-blue-500/20 z-10' : 'border-slate-200'}
                        `}
                        onClick={(e) => {
                            e.stopPropagation();
                            onSelectWidget(widget.id);
                        }}
                    >
                        {/* Header / Handle */}
                        <div className="drag-handle absolute top-0 left-0 right-0 h-6 cursor-move z-20 hover:bg-slate-50 rounded-t-lg" />

                        <div className="h-full w-full">
                            {/* pointer-events-none removed to allow chart interaction, handle handles drag */}
                            {renderWidgetContent(widget)}
                        </div>
                    </div>
                ))}
            </ResponsiveGridLayout>
        </div>
    );
};
