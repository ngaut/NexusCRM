import React from 'react';
import { AlertTriangle, X } from 'lucide-react';
import { Button } from '../ui/Button';

interface ConfirmationModalProps {
    isOpen: boolean;
    onClose: () => void;
    onConfirm: () => void;
    title: string;
    message: React.ReactNode;
    confirmLabel?: string;
    cancelLabel?: string;
    variant?: 'danger' | 'warning' | 'info';
    loading?: boolean;
    showCancel?: boolean;
    icon?: React.ReactNode;
}

export const ConfirmationModal: React.FC<ConfirmationModalProps> = ({
    isOpen,
    onClose,
    onConfirm,
    title,
    message,
    confirmLabel = 'Confirm',
    cancelLabel = 'Cancel',
    variant = 'danger',
    loading = false,
    showCancel = true,
    icon
}) => {
    if (!isOpen) {
        return null;
    }

    const getIcon = () => {
        if (icon) return icon;

        switch (variant) {
            case 'danger':
                return <AlertTriangle className="w-6 h-6 text-red-600" />;
            case 'warning':
                return <AlertTriangle className="w-6 h-6 text-yellow-600" />;
            default:
                return <AlertTriangle className="w-6 h-6 text-blue-600" />;
        }
    };

    const getHeaderBg = () => {
        switch (variant) {
            case 'danger':
                return 'bg-red-50';
            case 'warning':
                return 'bg-yellow-50';
            default:
                return 'bg-blue-50';
        }
    };

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50 backdrop-blur-sm animate-in fade-in duration-200">
            <div
                className="bg-white rounded-xl shadow-xl w-full max-w-md overflow-hidden transform transition-all scale-100"
                role="dialog"
                aria-modal="true"
                aria-labelledby="modal-title"
            >
                {/* Header */}
                <div className={`px-6 py-4 border-b border-gray-100 flex items-center justify-between ${getHeaderBg()}`}>
                    <div className="flex items-center gap-3">
                        {getIcon()}
                        <h3 className="text-lg font-semibold text-gray-900" id="modal-title">
                            {title}
                        </h3>
                    </div>
                    <button
                        onClick={onClose}
                        className="text-gray-400 hover:text-gray-500 transition-colors"
                        disabled={loading}
                    >
                        <X className="w-5 h-5" />
                    </button>
                </div>

                {/* Body */}
                <div className="p-6">
                    <div className="text-gray-600">
                        {message}
                    </div>
                </div>

                {/* Footer */}
                <div className="px-6 py-4 bg-gray-50 flex justify-end gap-3">
                    {showCancel && (
                        <Button
                            variant="ghost"
                            onClick={onClose}
                            disabled={loading}
                        >
                            {cancelLabel}
                        </Button>
                    )}
                    <Button
                        variant={variant === 'danger' ? 'danger' : 'primary'}
                        onClick={onConfirm}
                        loading={loading}
                    >
                        {confirmLabel}
                    </Button>
                </div>
            </div>
        </div>
    );
};
