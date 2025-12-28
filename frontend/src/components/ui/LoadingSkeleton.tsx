import React from 'react';

/**
 * Metadata-aware loading skeletons
 * Automatically adapts based on object metadata (field count, types)
 */

interface SkeletonProps {
    className?: string;
}

export function Skeleton({ className = '' }: SkeletonProps) {
    return (
        <div className={`animate-pulse bg-gray-200 rounded ${className}`} />
    );
}

// Record list skeleton - adapts to column count
export function RecordListSkeleton({
    rows = 5,
    columns = 6
}: {
    rows?: number;
    columns?: number;
}) {
    return (
        <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
            {/* Header */}
            <div className="border-b border-gray-200 px-6 py-3">
                <div className="flex gap-4">
                    {Array.from({ length: columns }).map((_, i) => (
                        <Skeleton key={i} className="h-4 w-24" />
                    ))}
                </div>
            </div>

            {/* Rows */}
            {Array.from({ length: rows }).map((_, rowIndex) => (
                <div key={rowIndex} className="border-b border-gray-100 px-6 py-4">
                    <div className="flex gap-4 items-center">
                        {Array.from({ length: columns }).map((_, colIndex) => (
                            <Skeleton
                                key={colIndex}
                                className={`h-4 ${colIndex === 0 ? 'w-32' : 'w-20'}`}
                            />
                        ))}
                    </div>
                </div>
            ))}
        </div>
    );
}

// Record detail skeleton - adapts to field count and layout
export function RecordDetailSkeleton({
    sections = 2,
    fieldsPerSection = 6
}: {
    sections?: number;
    fieldsPerSection?: number;
}) {
    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="bg-white rounded-lg border border-gray-200 p-6">
                <Skeleton className="h-8 w-64 mb-2" />
                <Skeleton className="h-4 w-48" />
            </div>

            {/* Sections */}
            {Array.from({ length: sections }).map((_, sectionIndex) => (
                <div key={sectionIndex} className="bg-white rounded-lg border border-gray-200 p-6">
                    <Skeleton className="h-6 w-32 mb-4" />
                    <div className="grid grid-cols-2 gap-4">
                        {Array.from({ length: fieldsPerSection }).map((_, fieldIndex) => (
                            <div key={fieldIndex}>
                                <Skeleton className="h-3 w-20 mb-2" />
                                <Skeleton className="h-5 w-full" />
                            </div>
                        ))}
                    </div>
                </div>
            ))}
        </div>
    );
}

// Form skeleton - adapts to field count
export function FormSkeleton({
    fields = 8
}: {
    fields?: number;
}) {
    return (
        <div className="space-y-4">
            {Array.from({ length: fields }).map((_, i) => (
                <div key={i}>
                    <Skeleton className="h-4 w-24 mb-2" />
                    <Skeleton className="h-10 w-full" />
                </div>
            ))}
        </div>
    );
}

// Card grid skeleton
export function CardGridSkeleton({
    count = 6
}: {
    count?: number;
}) {
    return (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {Array.from({ length: count }).map((_, i) => (
                <div key={i} className="bg-white rounded-lg border border-gray-200 p-6">
                    <Skeleton className="h-6 w-3/4 mb-4" />
                    <Skeleton className="h-4 w-full mb-2" />
                    <Skeleton className="h-4 w-2/3" />
                </div>
            ))}
        </div>
    );
}

// Dashboard widget skeleton
export function DashboardWidgetSkeleton() {
    return (
        <div className="bg-white rounded-lg border border-gray-200 p-6">
            <Skeleton className="h-5 w-32 mb-4" />
            <Skeleton className="h-12 w-24 mb-2" />
            <Skeleton className="h-3 w-48" />
        </div>
    );
}

// Metadata-driven skeleton generator
export function MetadataAwareSkeleton({
    fieldCount,
    layout = 'form'
}: {
    fieldCount: number;
    layout?: 'form' | 'detail' | 'list';
}) {
    if (layout === 'list') {
        return <RecordListSkeleton columns={Math.min(fieldCount, 8)} />;
    }

    if (layout === 'detail') {
        const sections = Math.ceil(fieldCount / 6);
        return <RecordDetailSkeleton sections={sections} />;
    }

    return <FormSkeleton fields={fieldCount} />;
}
