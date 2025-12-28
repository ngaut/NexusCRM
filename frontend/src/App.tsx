import React, { useState, Suspense } from 'react';
import { BrowserRouter } from 'react-router-dom';
import { Layout } from './components/Layout';
import { AIAssistant } from './components/AIAssistant';
import { NotificationProvider } from './contexts/NotificationContext';
import { ToastContainer } from './components/Toast';
import { RuntimeProvider, useRuntime } from './contexts/RuntimeContext';
import { AppProvider } from './contexts/AppContext';
import { Router } from './components/Router';
import { LoginScreen } from './components/auth/LoginScreen';
import { BootScreen } from './components/auth/BootScreen';
import { ErrorBoundary } from './components/ErrorBoundary';

import { ToastProvider } from './components/ui/Toast';
import { PermissionProvider } from './contexts/PermissionContext';

function AppContent() {

  const { authStatus, login } = useRuntime();
  const [isAIOpen, setIsAIOpen] = useState(false);
  const [loginError, setLoginError] = useState<string | null>(null);

  const handleLogin = async (email: string, password: string) => {
    if (!email || !password) {
      setLoginError("Please enter email and password.");
      return;
    }
    setLoginError(null);
    try {
      // runtime.login handles authentication and kernel boot
      await login(email, password);
    } catch (e: unknown) {
      const message = e instanceof Error ? e.message : "Login failed. Please check your credentials.";
      setLoginError(message);
    }
  };



  if (authStatus === 'loading') {
    return <BootScreen />;
  }

  if (authStatus === 'unauthenticated') {
    return (
      <LoginScreen
        onLogin={handleLogin}
        loginError={loginError}
      />
    );
  }

  return (
    <AppProvider>
      <PermissionProvider>
        <Router onToggleAI={() => setIsAIOpen(!isAIOpen)} />
        <AIAssistant
          isOpen={isAIOpen}
          onClose={() => setIsAIOpen(false)}
        />
        <ToastContainer />

      </PermissionProvider>
    </AppProvider>
  );
}

function App() {
  return (
    <ErrorBoundary>
      <NotificationProvider>
        <RuntimeProvider>
          <ToastProvider>
            <BrowserRouter>
              <AppContent />
            </BrowserRouter>
          </ToastProvider>
        </RuntimeProvider>
      </NotificationProvider>
    </ErrorBoundary>
  );
}

export default App;
