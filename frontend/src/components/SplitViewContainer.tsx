import React, { useState, useCallback, useRef, useEffect } from 'react';
import { Columns, Maximize2, GripVertical } from 'lucide-react';

// ============================================================================
// Types
// ============================================================================

interface SplitViewContainerProps {
    /** Content for the left panel (typically a list) */
    leftPanel: React.ReactNode;
    /** Content for the right panel (typically record detail) */
    rightPanel: React.ReactNode;
    /** Initial split position as percentage of left panel width (default: 40) */
    defaultSplit?: number;
    /** Minimum width of left panel in pixels (default: 300) */
    minLeftWidth?: number;
    /** Minimum width of right panel in pixels (default: 400) */
    minRightWidth?: number;
    /** Whether to show the right panel (controlled mode) */
    showRightPanel?: boolean;
    /** Callback when split view is toggled */
    onToggleSplit?: (isExpanded: boolean) => void;
}

// ============================================================================
// Main Component
// ============================================================================

/**
 * SplitViewContainer provides a resizable two-panel layout.
 * 
 * Features:
 * - Draggable divider for custom split position
 * - Minimum width constraints for both panels
 * - Collapse button to return to single-panel mode
 * - Controlled or uncontrolled mode via showRightPanel prop
 */
export const SplitViewContainer: React.FC<SplitViewContainerProps> = ({
    leftPanel,
    rightPanel,
    defaultSplit = 40,
    minLeftWidth = 300,
    minRightWidth = 400,
    showRightPanel = false,
    onToggleSplit
}) => {
    const [splitPosition, setSplitPosition] = useState(defaultSplit);
    const [isDragging, setIsDragging] = useState(false);
    const [isExpanded, setIsExpanded] = useState(showRightPanel);
    const containerRef = useRef<HTMLDivElement>(null);

    // Sync internal state with prop when controlled
    useEffect(() => {
        setIsExpanded(showRightPanel);
    }, [showRightPanel]);

    // ========== Drag Handling ==========

    const handleMouseDown = useCallback((e: React.MouseEvent) => {
        e.preventDefault();
        setIsDragging(true);
    }, []);

    const handleMouseMove = useCallback((e: MouseEvent) => {
        if (!isDragging || !containerRef.current) return;

        const containerRect = containerRef.current.getBoundingClientRect();
        const containerWidth = containerRect.width;
        const mouseX = e.clientX - containerRect.left;

        // Calculate new split as percentage
        let newSplit = (mouseX / containerWidth) * 100;

        // Enforce minimum widths
        const minLeftPercent = (minLeftWidth / containerWidth) * 100;
        const minRightPercent = (minRightWidth / containerWidth) * 100;
        newSplit = Math.max(minLeftPercent, Math.min(100 - minRightPercent, newSplit));

        setSplitPosition(newSplit);
    }, [isDragging, minLeftWidth, minRightWidth]);

    const handleMouseUp = useCallback(() => {
        setIsDragging(false);
    }, []);

    // Attach global mouse events during drag
    useEffect(() => {
        if (!isDragging) return;

        document.addEventListener('mousemove', handleMouseMove);
        document.addEventListener('mouseup', handleMouseUp);
        document.body.style.cursor = 'col-resize';
        document.body.style.userSelect = 'none';

        return () => {
            document.removeEventListener('mousemove', handleMouseMove);
            document.removeEventListener('mouseup', handleMouseUp);
            document.body.style.cursor = '';
            document.body.style.userSelect = '';
        };
    }, [isDragging, handleMouseMove, handleMouseUp]);

    // ========== Toggle Handler ==========

    const handleCollapse = useCallback(() => {
        setIsExpanded(false);
        onToggleSplit?.(false);
    }, [onToggleSplit]);

    // ========== Render ==========

    // Single-panel mode (no split)
    if (!isExpanded) {
        return (
            <div className="h-full flex flex-col">
                <div className="flex-1 overflow-auto">
                    {leftPanel}
                </div>
            </div>
        );
    }

    // Split mode
    return (
        <div ref={containerRef} className="h-full flex relative">
            {/* Left Panel */}
            <div
                className="h-full overflow-auto bg-white border-r border-gray-200"
                style={{ width: `${splitPosition}%` }}
            >
                {leftPanel}
            </div>

            {/* Resizer Handle */}
            <div
                onMouseDown={handleMouseDown}
                className={`
                    w-1 bg-gray-200 hover:bg-blue-500 cursor-col-resize
                    flex items-center justify-center transition-colors duration-150
                    ${isDragging ? 'bg-blue-500' : ''}
                `}
            >
                <div className="absolute flex items-center justify-center w-6 h-12 -ml-2.5 bg-gray-100 hover:bg-blue-100 rounded-full shadow border border-gray-200 transition-colors">
                    <GripVertical className="w-4 h-4 text-gray-400" />
                </div>
            </div>

            {/* Right Panel */}
            <div
                className="h-full overflow-auto bg-gray-50 flex-1"
                style={{ width: `${100 - splitPosition}%` }}
            >
                {rightPanel}
            </div>

            {/* Collapse Button */}
            <button
                onClick={handleCollapse}
                className="absolute top-4 right-4 z-10 p-2 bg-white rounded-lg shadow-md hover:bg-gray-50 transition-colors border border-gray-200"
                title="Exit split view"
            >
                <Maximize2 className="w-4 h-4 text-gray-600" />
            </button>
        </div>
    );
};

// ============================================================================
// Hook: useSplitView
// ============================================================================

interface UseSplitViewReturn {
    /** Currently selected record ID in split view */
    selectedRecordId: string | null;
    /** Whether split view is active */
    isSplitMode: boolean;
    /** Open a record in split view */
    openInSplit: (recordId: string) => void;
    /** Close split view */
    closeSplit: () => void;
    /** Toggle split view (closes if open) */
    toggleSplit: () => void;
    /** Set selected record without changing split mode */
    setSelectedRecordId: React.Dispatch<React.SetStateAction<string | null>>;
}

/**
 * Hook to manage split view state in parent components.
 * Provides methods to open, close, and toggle split view.
 */
export function useSplitView(): UseSplitViewReturn {
    const [selectedRecordId, setSelectedRecordId] = useState<string | null>(null);
    const [isSplitMode, setIsSplitMode] = useState(false);

    const openInSplit = useCallback((recordId: string) => {
        setSelectedRecordId(recordId);
        setIsSplitMode(true);
    }, []);

    const closeSplit = useCallback(() => {
        setSelectedRecordId(null);
        setIsSplitMode(false);
    }, []);

    const toggleSplit = useCallback(() => {
        setIsSplitMode(prev => {
            if (prev) {
                setSelectedRecordId(null);
            }
            return !prev;
        });
    }, []);

    return {
        selectedRecordId,
        isSplitMode,
        openInSplit,
        closeSplit,
        toggleSplit,
        setSelectedRecordId
    };
}

// ============================================================================
// Toggle Button Component
// ============================================================================

interface SplitViewToggleProps {
    /** Whether split view is currently enabled */
    isEnabled: boolean;
    /** Handler for toggle action */
    onToggle: () => void;
}

/**
 * Toggle button to enable/disable split view in toolbars.
 */
export const SplitViewToggle: React.FC<SplitViewToggleProps> = ({ isEnabled, onToggle }) => (
    <button
        onClick={onToggle}
        className={`
            p-2 rounded-lg border transition-all
            ${isEnabled
                ? 'bg-blue-50 border-blue-200 text-blue-600'
                : 'bg-white border-gray-200 text-gray-600 hover:bg-gray-50'
            }
        `}
        title={isEnabled ? 'Exit split view' : 'Enable split view'}
    >
        {isEnabled ? (
            <Maximize2 className="w-4 h-4" />
        ) : (
            <Columns className="w-4 h-4" />
        )}
    </button>
);

export default SplitViewContainer;
