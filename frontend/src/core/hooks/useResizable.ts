import { useState, useCallback, useRef, useEffect } from 'react';

interface UseResizableProps {
    initialWidth: number;
    minWidth: number;
    maxWidth: number;
    direction: 'left' | 'right'; // which side expands
}

export function useResizable({ initialWidth, minWidth, maxWidth, direction }: UseResizableProps) {
    const [width, setWidth] = useState(initialWidth);
    const isResizingRef = useRef(false);

    const startResizing = useCallback((e: React.MouseEvent) => {
        e.preventDefault();
        isResizingRef.current = true;
        const startX = e.clientX;
        const startWidth = width;

        const handleMouseMove = (moveEvent: MouseEvent) => {
            if (!isResizingRef.current) return;
            const delta = startX - moveEvent.clientX;
            // If dragging left edge (right panel expands leftwards): delta > 0 means width increases
            // If dragging right edge (left panel expands rightwards): delta > 0 means width decreases (usually we measure from left)

            // For right-sidebar (AIAssistant): dragging left edge left (startX > clientX) -> delta > 0 -> width increases.
            // For left-sidebar (ContextPanel inside AIAssistant): dragging left edge left -> delta > 0 -> width increases.

            const diff = direction === 'left' ? delta : -delta;

            const newWidth = Math.min(Math.max(startWidth + diff, minWidth), maxWidth);
            setWidth(newWidth);
        };

        const handleMouseUp = () => {
            isResizingRef.current = false;
            document.removeEventListener('mousemove', handleMouseMove);
            document.removeEventListener('mouseup', handleMouseUp);
            document.body.style.cursor = '';
            document.body.style.userSelect = '';
        };

        document.addEventListener('mousemove', handleMouseMove);
        document.addEventListener('mouseup', handleMouseUp);
        document.body.style.cursor = 'ew-resize';
        document.body.style.userSelect = 'none';
    }, [width, minWidth, maxWidth, direction]);

    return { width, setWidth, startResizing };
}
