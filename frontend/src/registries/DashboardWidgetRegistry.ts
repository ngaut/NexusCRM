
import React from 'react';
import { WidgetConfig } from '../types';
import { GlobalFilters } from '../components/FilterBar';
import { RegistryBase } from '@shared/utils';

// Import Widgets
// Import Widgets using shared props
import { MetricWidget, ChartWidget, PieWidget, FunnelWidget, GaugeWidget } from '../components/widgets/StandardWidgets';
import { RecordListWidget } from '../components/widgets/RecordListWidget';
import { KanbanWidget } from '../components/widgets/KanbanWidget';
import { SqlChartWidget } from '../components/widgets/SqlChartWidget';
import { WidgetRendererProps } from '../types';

class DashboardWidgetRegistryClass extends RegistryBase<React.FC<WidgetRendererProps>> {
    constructor() {
        super({
            validator: (key, value) => typeof value === 'function' ? true : `Widget "${key}" must be a component`,
            enableEvents: true
        });

        this.registerDefaults();
    }

    private registerDefaults() {
        // Standard Charts (from UIRegistry)
        this.registerWidget('metric', MetricWidget);
        this.registerWidget('chart-bar', ChartWidget);
        this.registerWidget('chart-line', ChartWidget);
        this.registerWidget('chart-pie', PieWidget); // Note: UIRegistry exports PieWidget alias
        this.registerWidget('chart-funnel', FunnelWidget);
        this.registerWidget('chart-gauge', GaugeWidget);

        // New Data Widgets
        this.registerWidget('record-list', RecordListWidget);
        this.registerWidget('kanban', KanbanWidget);
        this.registerWidget('sql-chart', SqlChartWidget);
    }

    registerWidget(type: string, component: React.FC<WidgetRendererProps>) {
        this.register(type, component);
    }

    getWidget(type: string): React.FC<WidgetRendererProps> | null {
        return this.get(type) || null;
    }
}

export const DashboardWidgetRegistry = new DashboardWidgetRegistryClass();
