import React, { createContext, useContext, useState, useCallback } from 'react';
import { X, CheckCircle, AlertCircle, XCircle, Info } from 'lucide-react';

type ToastType = 'success' | 'error' | 'warning' | 'info';

interface Toast {
    id: string;
    type: ToastType;
    title: string;
    message?: string;
    duration?: number;
}

interface ToastContextType {
    showToast: (toast: Omit<Toast, 'id'>) => void;
    hideToast: (id: string) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

export function useToast() {
    const context = useContext(ToastContext);
    if (!context) {
        throw new Error('useToast must be used within ToastProvider');
    }
    return context;
}

export function ToastProvider({ children }: { children: React.ReactNode }) {
    const [toasts, setToasts] = useState<Toast[]>([]);

    const showToast = useCallback((toast: Omit<Toast, 'id'>) => {
        const id = Math.random().toString(36).substring(7);
        const newToast: Toast = { ...toast, id };

        setToasts((prev) => [...prev, newToast]);

        // Auto-dismiss after duration
        const duration = toast.duration ?? 5000;
        if (duration > 0) {
            setTimeout(() => {
                hideToast(id);
            }, duration);
        }
    }, []);

    const hideToast = useCallback((id: string) => {
        setToasts((prev) => prev.filter((toast) => toast.id !== id));
    }, []);

    return (
        <ToastContext.Provider value={{ showToast, hideToast }}>
            {children}
            <ToastContainer toasts={toasts} onDismiss={hideToast} />
        </ToastContext.Provider>
    );
}

function ToastContainer({ toasts, onDismiss }: { toasts: Toast[]; onDismiss: (id: string) => void }) {
    return (
        <div className="fixed top-4 right-4 z-50 flex flex-col gap-2 pointer-events-none">
            {toasts.map((toast) => (
                <ToastItem key={toast.id} toast={toast} onDismiss={onDismiss} />
            ))}
        </div>
    );
}

const toastConfig: Record<ToastType, { icon: typeof CheckCircle; bgColor: string; iconColor: string; borderColor: string }> = {
    success: {
        icon: CheckCircle,
        bgColor: 'bg-green-50',
        iconColor: 'text-green-600',
        borderColor: 'border-green-200',
    },
    error: {
        icon: XCircle,
        bgColor: 'bg-red-50',
        iconColor: 'text-red-600',
        borderColor: 'border-red-200',
    },
    warning: {
        icon: AlertCircle,
        bgColor: 'bg-yellow-50',
        iconColor: 'text-yellow-600',
        borderColor: 'border-yellow-200',
    },
    info: {
        icon: Info,
        bgColor: 'bg-blue-50',
        iconColor: 'text-blue-600',
        borderColor: 'border-blue-200',
    },
};

function ToastItem({ toast, onDismiss }: { toast: Toast; onDismiss: (id: string) => void }) {
    const config = toastConfig[toast.type];
    const Icon = config.icon;

    return (
        <div
            className={`
        pointer-events-auto min-w-[300px] max-w-md
        ${config.bgColor} ${config.borderColor}
        border rounded-lg shadow-lg p-4
        animate-slide-in-right
      `}
        >
            <div className="flex gap-3">
                <Icon className={`w-5 h-5 ${config.iconColor} flex-shrink-0 mt-0.5`} />
                <div className="flex-1">
                    <h4 className="text-sm font-semibold text-gray-900">{toast.title}</h4>
                    {toast.message && (
                        <p className="text-sm text-gray-600 mt-1">{toast.message}</p>
                    )}
                </div>
                <button
                    onClick={() => onDismiss(toast.id)}
                    className="text-gray-400 hover:text-gray-600 transition-colors"
                >
                    <X className="w-4 h-4" />
                </button>
            </div>
        </div>
    );
}

// Convenience hooks for common toast types
export function useSuccessToast() {
    const { showToast } = useToast();
    return useCallback((title: string, message?: string) => {
        showToast({ type: 'success', title, message });
    }, [showToast]);
}

export function useErrorToast() {
    const { showToast } = useToast();
    return useCallback((title: string, message?: string) => {
        showToast({ type: 'error', title, message, duration: 10000 });
    }, [showToast]);
}

export function useInfoToast() {
    const { showToast } = useToast();
    return useCallback((title: string, message?: string) => {
        showToast({ type: 'info', title, message });
    }, [showToast]);
}
