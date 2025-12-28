import React, { useEffect, useRef } from 'react';
import { X, LucideIcon } from 'lucide-react';

export interface BaseModalProps {
    isOpen: boolean;
    onClose: () => void;
    title: React.ReactNode;
    description?: React.ReactNode;
    icon?: LucideIcon;
    iconClassName?: string;
    iconBgClassName?: string;
    headerClassName?: string;
    maxWidth?: 'sm' | 'md' | 'lg' | 'xl' | '2xl' | '3xl' | '4xl' | '5xl' | 'full';
    children: React.ReactNode;
    footer?: React.ReactNode;
    closeButtonLight?: boolean; // For dark headers
}

const MAX_WIDTHS = {
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-lg',
    xl: 'max-w-xl',
    '2xl': 'max-w-2xl',
    '3xl': 'max-w-3xl',
    '4xl': 'max-w-4xl',
    '5xl': 'max-w-5xl',
    'full': 'max-w-full mx-4',
};

export function BaseModal({
    isOpen,
    onClose,
    title,
    description,
    icon: Icon,
    iconClassName = 'text-gray-600',
    iconBgClassName = 'bg-gray-100',
    headerClassName = 'bg-white border-b border-gray-100',
    maxWidth = 'lg',
    children,
    footer,
    closeButtonLight = false
}: BaseModalProps) {
    const modalRef = useRef<HTMLDivElement>(null);

    // Close on escape key
    useEffect(() => {
        const handleEscape = (e: KeyboardEvent) => {
            if (isOpen && e.key === 'Escape') {
                onClose();
            }
        };
        document.addEventListener('keydown', handleEscape);
        return () => document.removeEventListener('keydown', handleEscape);
    }, [isOpen, onClose]);

    // Focus trap could be added here

    if (!isOpen) return null;

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
            {/* Backdrop */}
            <div
                className="absolute inset-0 bg-black/50 backdrop-blur-sm transition-opacity"
                onClick={onClose}
                aria-hidden="true"
            />

            {/* Modal Container */}
            <div
                ref={modalRef}
                className={`relative bg-white rounded-2xl shadow-2xl w-full ${MAX_WIDTHS[maxWidth]} overflow-hidden animate-fade-in flex flex-col max-h-[90vh]`}
                role="dialog"
                aria-modal="true"
            >
                {/* Header */}
                <div className={`flex items-center justify-between px-6 py-4 flex-shrink-0 ${headerClassName}`}>
                    <div className="flex items-center gap-3">
                        {Icon && (
                            <div className={`p-2 rounded-lg flex-shrink-0 ${iconBgClassName}`}>
                                <Icon className={`w-6 h-6 ${iconClassName}`} />
                            </div>
                        )}
                        <div>
                            <h2 className={`text-xl font-bold ${closeButtonLight ? 'text-white' : 'text-gray-900'}`}>
                                {title}
                            </h2>
                            {description && (
                                <div className={`text-sm ${closeButtonLight ? 'text-white/80' : 'text-gray-500'}`}>
                                    {description}
                                </div>
                            )}
                        </div>
                    </div>
                    <button
                        onClick={onClose}
                        className={`p-2 rounded-full transition-colors ${closeButtonLight
                            ? 'hover:bg-white/20 text-white/80 hover:text-white'
                            : 'hover:bg-gray-100 text-gray-500 hover:text-gray-700'
                            }`}
                        aria-label="Close modal"
                    >
                        <X className="w-5 h-5" />
                    </button>
                </div>

                {/* Content */}
                <div className="flex-1 overflow-y-auto">
                    {children}
                </div>

                {/* Footer */}
                {footer && (
                    <div className="flex-shrink-0 border-t border-gray-100 bg-gray-50 px-6 py-4">
                        {footer}
                    </div>
                )}
            </div>
        </div>
    );
}
