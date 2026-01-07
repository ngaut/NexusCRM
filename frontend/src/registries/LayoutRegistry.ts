import { RegistryBase } from '@shared/utils';
import { PageLayout, LayoutType } from '../types';
import { metadataAPI } from '../infrastructure/api/metadata';
import { Logger } from '../core/services/Logger';

export class LayoutRegistryClass extends RegistryBase<PageLayout> {
    constructor() {
        super({
            validator: (key, value) => !!value.object_api_name && !!value.layout_name && !!value.sections,
            enableEvents: true
        });
    }

    /**
     * Retrieves a layout for a specific object and type.
     * Falls back to a default layout if the specific one isn't found.
     */
    public async getLayout(objectName: string, type: LayoutType = 'Detail'): Promise<PageLayout> {
        // 1. Try to find an in-memory override first (registered via code)
        const memoryKey = `${objectName}:${type}`;
        if (this.has(memoryKey)) {
            return this.get(memoryKey)!;
        }

        // 2. Try to fetch from backend API
        try {
            // Note: Currently only fetching default layout by object name. 
            // Future: Support profile-specific layouts via typed params if needed.
            const response = await metadataAPI.getLayout(objectName);
            if (response && response.layout) {
                // Cache it? Ideally yes, but for now just return
                return response.layout;
            }
        } catch (error) {
            Logger.warn(`Failed to load layout for ${objectName} (${type}) from API, falling back to default.`, error);
        }

        // 3. Generate a default layout if nothing exists
        return this.generateDefaultLayout(objectName, type);
    }

    /**
     * Generates a simple default layout based on the object's fields.
     */
    private generateDefaultLayout(objectName: string, type: LayoutType): PageLayout {
        // For now, return a minimal default layout
        return {
            id: 'default',
            object_api_name: objectName,
            layout_name: `Default ${type} Layout`,
            type,
            is_default: true,
            sections: [
                {
                    id: 'default_section',
                    label: 'General Information',
                    columns: 2,
                    fields: [] // Will be populated when API client is implemented
                }
            ],
            compact_layout: [],
            related_lists: [],
            header_actions: [],
            quick_actions: []
        };
    }

    /**
     * Saves a layout to the backend.
     */
    public async saveLayout(layout: PageLayout): Promise<void> {
        try {
            await metadataAPI.saveLayout(layout);
        } catch (error) {
            Logger.error('Failed to save layout via API', error);
            throw error;
        }

        // Update in-memory cache
        this.register(`${layout.object_api_name}:${layout.type}`, layout);
    }
}

export const LayoutRegistry = new LayoutRegistryClass();
