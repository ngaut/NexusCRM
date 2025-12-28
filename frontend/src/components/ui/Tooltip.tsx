import React, { useState } from 'react';

interface TooltipProps {
    content: string;
    children: React.ReactNode;
    position?: 'right' | 'top' | 'bottom' | 'left';
    delay?: number;
}

export const Tooltip: React.FC<TooltipProps> = ({
    content,
    children,
    position = 'right',
    delay = 200
}) => {
    const [isVisible, setIsVisible] = useState(false);
    const [timeoutId, setTimeoutId] = useState<NodeJS.Timeout | null>(null);

    const handleMouseEnter = () => {
        const id = setTimeout(() => setIsVisible(true), delay);
        setTimeoutId(id);
    };

    const handleMouseLeave = () => {
        if (timeoutId) clearTimeout(timeoutId);
        setIsVisible(false);
    };

    return (
        <div
            className="relative flex items-center"
            onMouseEnter={handleMouseEnter}
            onMouseLeave={handleMouseLeave}
        >
            {children}
            {isVisible && (
                <div className={`absolute z-50 px-2 py-1 text-xs font-medium text-white bg-slate-900 rounded shadow-lg whitespace-nowrap animate-in fade-in zoom-in-95 duration-150 ${position === 'right' ? 'left-full ml-2' :
                        position === 'left' ? 'right-full mr-2' :
                            position === 'top' ? 'bottom-full mb-2' :
                                'top-full mt-2'
                    }`}>
                    {content}
                    {/* Arrow */}
                    <div className={`absolute w-0 h-0 border-4 border-transparent ${position === 'right' ? 'border-r-slate-900 right-full top-1/2 -translate-y-1/2 -mr-1' :
                            position === 'left' ? 'border-l-slate-900 left-full top-1/2 -translate-y-1/2 -ml-1' :
                                position === 'top' ? 'border-t-slate-900 top-full left-1/2 -translate-x-1/2 -mt-1' :
                                    'border-b-slate-900 bottom-full left-1/2 -translate-x-1/2 -mb-1'
                        }`} />
                </div>
            )}
        </div>
    );
};
