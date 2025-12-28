import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';

import { MetadataRegistry } from './registries/MetadataRegistry';
import { registerPlugins } from './registerPlugins';

// Initialize plugins
registerPlugins();
// Initialize dynamic field types from plugins
MetadataRegistry.loadDynamicTypes();

import { ThemeProvider } from './contexts/ThemeContext';

const rootElement = document.getElementById('root');
if (!rootElement) {
  throw new Error("Could not find root element to mount to");
}

const root = ReactDOM.createRoot(rootElement);
root.render(
  <React.StrictMode>
    <ThemeProvider>
      <App />
    </ThemeProvider>
  </React.StrictMode>
);