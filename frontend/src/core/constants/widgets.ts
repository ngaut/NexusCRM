export const WIDGET_SIZE_DEFAULTS: Record<string, { w: number, h: number }> = {
    'metric': { w: 3, h: 2 },
    'chart-bar': { w: 4, h: 3 },
    'chart-line': { w: 4, h: 3 },
    'chart-pie': { w: 3, h: 3 },
    'chart-funnel': { w: 4, h: 3 },
    'chart-gauge': { w: 3, h: 3 },
    'record-list': { w: 6, h: 4 },
    'kanban': { w: 12, h: 5 },
    'sql-chart': { w: 6, h: 4 },
    'text': { w: 4, h: 2 },
    'image': { w: 3, h: 3 }
};

export const DEFAULT_WIDGET_SIZE = { w: 4, h: 3 };
