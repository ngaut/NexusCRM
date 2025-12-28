import React, { createContext, useContext, useEffect, useState } from 'react';
import { metadataAPI } from '../infrastructure/api/metadata';
import type { Theme } from '../types';

interface ThemeContextType {
    theme: Theme | null;
    isLoading: boolean;
}

const ThemeContext = createContext<ThemeContextType>({
    theme: null,
    isLoading: true,
});

const DEFAULT_THEME: Theme = {
    id: 'default',
    name: 'Default',
    colors: {
        brand: '#4F46E5',
        brand_light: '#818CF8',
        brand_dark: '#3730A3',
        secondary: '#64748B',
        success: '#10B981',
        warning: '#F59E0B',
        danger: '#EF4444',
        background: '#F1F5F9',
        surface: '#FFFFFF',
        text: '#0F172A',
        text_secondary: '#64748B',
        border: '#E2E8F0',
    },
    density: 'comfortable',
};

export const ThemeProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const [theme, setTheme] = useState<Theme | null>(null);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        const fetchTheme = async () => {
            try {
                // Fetch active theme from backend
                const response = await metadataAPI.getActiveTheme();
                if (response) {
                    applyTheme(response);
                    setTheme(response);
                } else {
                    applyTheme(DEFAULT_THEME);
                    setTheme(DEFAULT_THEME);
                }
            } catch (error) {
                console.warn('Failed to fetch theme, using default:', error);
                applyTheme(DEFAULT_THEME);
                setTheme(DEFAULT_THEME);
            } finally {
                setIsLoading(false);
            }
        };

        fetchTheme();
    }, []);

    const applyTheme = (theme: Theme) => {
        const root = document.documentElement;
        Object.entries(theme.colors).forEach(([key, value]) => {
            root.style.setProperty(`--color-${key}`, value as string);
        });
        root.style.setProperty('--density', theme.density === 'compact' ? '0.75rem' : '1rem');
    };

    return (
        <ThemeContext.Provider value={{ theme, isLoading }}>
            {children}
        </ThemeContext.Provider>
    );
};

export const useTheme = () => useContext(ThemeContext);
