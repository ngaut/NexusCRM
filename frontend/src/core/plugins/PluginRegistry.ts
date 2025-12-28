
import { Plugin, PluginType, ActionPlugin } from './types';

class PluginRegistry {
    private plugins: Map<string, Plugin> = new Map();

    register(plugin: Plugin) {
        if (this.plugins.has(plugin.name)) {
            console.warn(`Plugin '${plugin.name}' already registered. Overwriting.`);
        }

        this.plugins.set(plugin.name, plugin);
    }

    get(name: string): Plugin | undefined {
        return this.plugins.get(name);
    }

    getAction(name: string): ActionPlugin | undefined {
        const plugin = this.plugins.get(name);
        if (plugin?.type === PluginType.ACTION) {
            return plugin as ActionPlugin;
        }
        return undefined;
    }

    getAll(): Plugin[] {
        return Array.from(this.plugins.values());
    }
}

export const pluginRegistry = new PluginRegistry();
