import React, { createContext, useContext, useState, useEffect, useRef, ReactNode } from 'react';
import { authAPI, apiClient } from '../infrastructure/api';
import { authEvents, AUTH_EVENT_UNAUTHORIZED } from '../infrastructure/api/client';
import type { UserSession } from '../types';

interface RuntimeContextValue {
  authStatus: 'loading' | 'authenticated' | 'unauthenticated';
  user: UserSession | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  activeTabId: string | null;
  setActiveTabId: (tabId: string | null) => void;
}

const RuntimeContext = createContext<RuntimeContextValue | null>(null);

export function RuntimeProvider({ children }: { children: ReactNode }) {
  const [authStatus, setAuthStatus] = useState<'loading' | 'authenticated' | 'unauthenticated'>('loading');
  const [user, setUser] = useState<UserSession | null>(null);
  const [activeTabId, setActiveTabId] = useState<string | null>(null);

  // Track if user has ever been authenticated in this session.
  // This ref is used to distinguish between:
  // 1. Initial boot with invalid token (expected, just show login)
  // 2. Mid-session 401 (unexpected, force logout)
  const wasAuthenticatedRef = useRef(false);

  // Check for existing session on mount
  useEffect(() => {
    const checkAuth = async () => {
      if (authAPI.isAuthenticated()) {
        try {
          const userData = await authAPI.verify();
          setUser(userData);
          setAuthStatus('authenticated');
          wasAuthenticatedRef.current = true;
        } catch {
          // Token expired or invalid during boot - this is expected
          // Just clear state and show login, no need for special handling
          apiClient.setToken(null);
          setAuthStatus('unauthenticated');
          setUser(null);
        }
      } else {
        setAuthStatus('unauthenticated');
      }
    };

    checkAuth();

    // Listen for unauthorized events globally
    // This only processes events when user WAS authenticated (mid-session expiry)
    const handleUnauthorized = () => {
      if (wasAuthenticatedRef.current) {
        // User was authenticated but session expired mid-use
        wasAuthenticatedRef.current = false;
        setAuthStatus('unauthenticated');
        setUser(null);
        apiClient.setToken(null);
      }
      // If wasAuthenticatedRef is false, this is a boot-time 401 - ignore it
      // The checkAuth() flow already handles this case
    };

    authEvents.addEventListener(AUTH_EVENT_UNAUTHORIZED, handleUnauthorized);

    return () => {
      authEvents.removeEventListener(AUTH_EVENT_UNAUTHORIZED, handleUnauthorized);
    };
  }, []);

  const login = async (email: string, password: string) => {
    setAuthStatus('loading');
    try {
      const response = await authAPI.login({ email, password });
      setUser(response.user);
      setAuthStatus('authenticated');
      wasAuthenticatedRef.current = true;
    } catch (error) {
      setAuthStatus('unauthenticated');
      setUser(null);
      throw error;
    }
  };

  const logout = async () => {
    try {
      await authAPI.logout();
    } catch (error) {
      console.error('Logout failed:', error instanceof Error ? error.message : 'Unknown error');
    } finally {
      wasAuthenticatedRef.current = false;
      setUser(null);
      setAuthStatus('unauthenticated');
      setActiveTabId(null);
    }
  };

  return (
    <RuntimeContext.Provider value={{ authStatus, user, login, logout, activeTabId, setActiveTabId }}>
      {children}
    </RuntimeContext.Provider>
  );
}

export function useRuntime() {
  const context = useContext(RuntimeContext);
  if (!context) {
    throw new Error('useRuntime must be used within RuntimeProvider');
  }
  return context;
}
