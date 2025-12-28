import React from 'react';
import { RecordFieldsTab, RelatedListsTab, RecordFeedTab, TabProps } from '../components/tabs/RecordTabs';
import { ComponentRegistry } from './ComponentRegistry';
import { RegistryBase } from '@shared/utils';

class TabRegistryClass extends RegistryBase<React.FC<TabProps>> {
    constructor() {
        super({
            validator: (key, value) => typeof value === 'function' || typeof value === 'object' ? true : `Tab "${key}" must be a component`,
            enableEvents: true
        });
        this.registerDefaults();
    }

    private registerDefaults() {
        // Standard Tab Mappings
        this.register('Details', RecordFieldsTab);
        this.register('Info', RecordFieldsTab);
        this.register('Related', RelatedListsTab);
        this.register('Lists', RelatedListsTab);
        this.register('Feed', RecordFeedTab);
        this.register('Chatter', RecordFeedTab);
        this.register('Activity', RecordFeedTab);
    }

    /**
     * Resolves a Tab Name to a Component.
     * 1. Checks TabRegistry
     * 2. Checks ComponentRegistry (for custom components used as tabs)
     */
    get(name: string): React.FC<TabProps> | null {
        const tabComp = super.get(name);
        if (tabComp) return tabComp;

        // Fallback: Allow generic components to be used as tabs
        const genericComp = ComponentRegistry.get(name);
        if (genericComp) {
            // Wrap generic component to match TabProps signature if needed,
            // but generic components usually accept `record` prop which is compatible.
            return genericComp as React.FC<TabProps>;
        }

        return null;
    }
}

export const TabRegistry = new TabRegistryClass();
