import React, { Component, ErrorInfo, ReactNode } from 'react';
import { AlertTriangle, RefreshCw, Home, ChevronDown, ChevronUp } from 'lucide-react';

interface ErrorBoundaryProps {
    children: ReactNode;
    fallback?: ReactNode;
    onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

interface ErrorBoundaryState {
    hasError: boolean;
    error: Error | null;
    errorInfo: ErrorInfo | null;
    showDetails: boolean;
}

/**
 * Production-ready Error Boundary component
 * Catches JavaScript errors anywhere in the child component tree
 */
export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
    constructor(props: ErrorBoundaryProps) {
        super(props);
        this.state = {
            hasError: false,
            error: null,
            errorInfo: null,
            showDetails: false
        };
    }

    static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
        // Update state so the next render will show the fallback UI
        return { hasError: true };
    }

    componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
        // Log error to console in development
        if ((import.meta as unknown as { env: { DEV: boolean } }).env.DEV) {
            console.error('ErrorBoundary caught an error:', error);
            console.error('Component stack:', errorInfo.componentStack);
        }

        // Update state with error details
        this.setState({
            error,
            errorInfo
        });

        // Call custom error handler if provided
        if (this.props.onError) {
            this.props.onError(error, errorInfo);
        }

        // In production, you might want to send to an error reporting service
        // Example: Sentry.captureException(error, { extra: errorInfo });
    }

    handleReset = (): void => {
        this.setState({
            hasError: false,
            error: null,
            errorInfo: null,
            showDetails: false
        });
    };

    handleReload = (): void => {
        window.location.reload();
    };

    handleGoHome = (): void => {
        window.location.href = '/';
    };

    toggleDetails = (): void => {
        this.setState(prev => ({ showDetails: !prev.showDetails }));
    };

    render(): ReactNode {
        if (this.state.hasError) {
            // Use custom fallback if provided
            if (this.props.fallback) {
                return this.props.fallback;
            }

            // Default error UI
            return (
                <div className="min-h-screen bg-slate-50 flex items-center justify-center p-6">
                    <div className="max-w-2xl w-full bg-white rounded-lg shadow-lg border border-slate-200 overflow-hidden">
                        {/* Header */}
                        <div className="bg-red-50 border-b border-red-200 px-6 py-4">
                            <div className="flex items-center gap-3">
                                <div className="p-2 bg-red-100 rounded-full">
                                    <AlertTriangle className="text-red-600" size={24} />
                                </div>
                                <div>
                                    <h1 className="text-xl font-bold text-red-900">
                                        Something went wrong
                                    </h1>
                                    <p className="text-sm text-red-700 mt-0.5">
                                        An unexpected error occurred in the application
                                    </p>
                                </div>
                            </div>
                        </div>

                        {/* Error Message */}
                        <div className="px-6 py-4 border-b border-slate-200">
                            <h2 className="text-sm font-semibold text-slate-700 mb-2">
                                Error Details:
                            </h2>
                            <div className="bg-slate-50 border border-slate-200 rounded p-3">
                                <p className="text-sm text-slate-900 font-mono">
                                    {this.state.error?.toString() || 'Unknown error'}
                                </p>
                            </div>
                        </div>

                        {/* Stack Trace (Development Only) */}
                        {(import.meta as unknown as { env: { DEV: boolean } }).env.DEV && this.state.errorInfo && (
                            <div className="px-6 py-4 border-b border-slate-200">
                                <button
                                    onClick={this.toggleDetails}
                                    className="flex items-center gap-2 text-sm font-semibold text-slate-700 hover:text-slate-900 transition-colors"
                                >
                                    {this.state.showDetails ? (
                                        <ChevronUp size={16} />
                                    ) : (
                                        <ChevronDown size={16} />
                                    )}
                                    {this.state.showDetails ? 'Hide' : 'Show'} Component Stack
                                </button>

                                {this.state.showDetails && (
                                    <div className="mt-3 bg-slate-900 rounded p-4 overflow-x-auto">
                                        <pre className="text-xs text-slate-100 font-mono whitespace-pre-wrap">
                                            {this.state.errorInfo.componentStack}
                                        </pre>
                                    </div>
                                )}
                            </div>
                        )}

                        {/* Actions */}
                        <div className="px-6 py-4 bg-slate-50">
                            <div className="flex flex-wrap items-center gap-3">
                                <button
                                    onClick={this.handleReset}
                                    className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 transition-colors"
                                >
                                    <RefreshCw size={16} />
                                    Try Again
                                </button>

                                <button
                                    onClick={this.handleReload}
                                    className="flex items-center gap-2 px-4 py-2 bg-slate-600 text-white text-sm font-medium rounded-lg hover:bg-slate-700 transition-colors"
                                >
                                    <RefreshCw size={16} />
                                    Reload Page
                                </button>

                                <button
                                    onClick={this.handleGoHome}
                                    className="flex items-center gap-2 px-4 py-2 bg-white text-slate-700 text-sm font-medium border border-slate-300 rounded-lg hover:bg-slate-50 transition-colors"
                                >
                                    <Home size={16} />
                                    Go Home
                                </button>
                            </div>

                            <p className="text-xs text-slate-500 mt-4">
                                If this problem persists, please contact support or check the browser console for more details.
                            </p>
                        </div>
                    </div>
                </div>
            );
        }

        return this.props.children;
    }
}

/**
 * Hook-style wrapper for functional components
 * Usage: Wrap your component tree with <ErrorBoundary>
 */
export function withErrorBoundary<P extends object>(
    Component: React.ComponentType<P>,
    errorBoundaryProps?: Omit<ErrorBoundaryProps, 'children'>
): React.FC<P> {
    return (props: P) => (
        <ErrorBoundary {...errorBoundaryProps}>
            <Component {...props} />
        </ErrorBoundary>
    );
}
