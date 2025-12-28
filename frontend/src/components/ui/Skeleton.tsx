import React from 'react';

interface SkeletonProps {
    className?: string;
    width?: string | number;
    height?: string | number;
    variant?: 'text' | 'rectangular' | 'circular';
}

export const Skeleton: React.FC<SkeletonProps> = ({
    className = '',
    width,
    height,
    variant = 'rectangular'
}) => {
    const baseClasses = "animate-pulse bg-slate-200";
    const variantClasses = {
        text: "rounded-md",
        rectangular: "rounded-lg",
        circular: "rounded-full"
    };

    const style = {
        width: width,
        height: height
    };

    // Default heights for text variant if not specified
    if (variant === 'text' && !height) {
        style.height = '1em';
    }

    return (
        <div
            className={`${baseClasses} ${variantClasses[variant]} ${className}`}
            style={style}
        />
    );
};
