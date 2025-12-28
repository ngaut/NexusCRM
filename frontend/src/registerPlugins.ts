
import { pluginRegistry } from './core/plugins/PluginRegistry';
import { LeadConvertPlugin } from './plugins/actions/LeadConvertPlugin';

// Register all plugins here
export const registerPlugins = () => {
    pluginRegistry.register(LeadConvertPlugin);

    // Future plugins can be auto-imported or added here

};
