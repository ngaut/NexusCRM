import React from 'react';
import {
    FileText,
    Plus,
    Search,
    AlertCircle,
    CheckCircle,
    XCircle,
    Database,
    LucideIcon
} from 'lucide-react';
import { Button } from './Button';

export type EmptyStateVariant = 'noData' | 'noResults' | 'error' | 'success' | 'noAccess';

interface EmptyStateProps {
    variant?: EmptyStateVariant;
    icon?: LucideIcon;
    title: string;
    description?: string;
    action?: {
        label: string;
        onClick: () => void;
        icon?: React.ReactNode;
    };
    secondaryAction?: {
        label: string;
        onClick: () => void;
    };
}

const variantConfig: Record<EmptyStateVariant, {
    icon: LucideIcon;
    iconColor: string;
    iconBgColor: string;
}> = {
    noData: {
        icon: Database,
        iconColor: 'text-blue-600',
        iconBgColor: 'bg-blue-100',
    },
    noResults: {
        icon: Search,
        iconColor: 'text-gray-600',
        iconBgColor: 'bg-gray-100',
    },
    error: {
        icon: XCircle,
        iconColor: 'text-red-600',
        iconBgColor: 'bg-red-100',
    },
    success: {
        icon: CheckCircle,
        iconColor: 'text-green-600',
        iconBgColor: 'bg-green-100',
    },
    noAccess: {
        icon: AlertCircle,
        iconColor: 'text-yellow-600',
        iconBgColor: 'bg-yellow-100',
    },
};

export function EmptyState({
    variant = 'noData',
    icon: CustomIcon,
    title,
    description,
    action,
    secondaryAction,
}: EmptyStateProps) {
    const config = variantConfig[variant];
    const IconComponent = CustomIcon || config.icon;

    return (
        <div className="flex flex-col items-center justify-center py-12 px-4 text-center">
            {/* Icon */}
            <div className={`
        w-16 h-16 rounded-full flex items-center justify-center mb-4
        ${config.iconBgColor}
      `}>
                <IconComponent className={`w-8 h-8 ${config.iconColor}`} />
            </div>

            {/* Title */}
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
                {title}
            </h3>

            {/* Description */}
            {description && (
                <p className="text-sm text-gray-600 max-w-md mb-6">
                    {description}
                </p>
            )}

            {/* Actions */}
            {(action || secondaryAction) && (
                <div className="flex items-center gap-3">
                    {action && (
                        <Button
                            variant="primary"
                            onClick={action.onClick}
                            icon={action.icon}
                        >
                            {action.label}
                        </Button>
                    )}
                    {secondaryAction && (
                        <Button
                            variant="ghost"
                            onClick={secondaryAction.onClick}
                        >
                            {secondaryAction.label}
                        </Button>
                    )}
                </div>
            )}
        </div>
    );
}

// Preset empty states for common scenarios
export function NoRecordsEmptyState({
    objectName,
    onCreateNew
}: {
    objectName: string;
    onCreateNew: () => void;
}) {
    return (
        <EmptyState
            variant="noData"
            title={`No ${objectName} yet`}
            description={`Get started by creating your first ${objectName.toLowerCase()}. You can add details and track information here.`}
            action={{
                label: `New ${objectName}`,
                onClick: onCreateNew,
                icon: <Plus className="w-4 h-4" />,
            }}
        />
    );
}

export function NoSearchResultsEmptyState({
    searchTerm,
    onClearSearch
}: {
    searchTerm: string;
    onClearSearch: () => void;
}) {
    return (
        <EmptyState
            variant="noResults"
            title="No results found"
            description={`We couldn't find any records matching "${searchTerm}". Try adjusting your search or filters.`}
            action={{
                label: 'Clear Search',
                onClick: onClearSearch,
            }}
        />
    );
}

export function ErrorEmptyState({
    onRetry
}: {
    onRetry: () => void;
}) {
    return (
        <EmptyState
            variant="error"
            title="Something went wrong"
            description="We encountered an error loading this data. Please try again."
            action={{
                label: 'Retry',
                onClick: onRetry,
            }}
        />
    );
}

export function AccessDeniedEmptyState({
    onGoBack
}: {
    onGoBack: () => void;
}) {
    return (
        <EmptyState
            variant="noAccess"
            title="Access Denied"
            description="You don't have permission to view this record. This may be because the owner has changed or you're not part of the sharing group."
            action={{
                label: 'Go Back',
                onClick: onGoBack,
            }}
        />
    );
}
