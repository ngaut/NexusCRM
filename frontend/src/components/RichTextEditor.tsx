import React from 'react';
import { useEditor, EditorContent, ReactRenderer, Editor } from '@tiptap/react';
import StarterKit from '@tiptap/starter-kit';
import Mention from '@tiptap/extension-mention';
import tippy from 'tippy.js';
import 'tippy.js/dist/tippy.css';
import { SYSTEM_PROFILES } from '../core/constants/SystemProfiles';
import { KEYS } from '../core/constants';

interface RichTextEditorProps {
    value: string;
    onChange: (html: string) => void;
    placeholder?: string;
    mentionItems?: SuggestionItem[];
}

// Mock user list for mentions.
interface SuggestionItem {
    id: string;
    label: string;
}
interface Range {
    from: number;
    to: number;
}

interface SuggestionProps {
    editor: Editor;
    range: Range;
    query: string;
    text: string;
    items: SuggestionItem[];
    command: (props: { id: string; label: string }) => void;
    decorationNode: Element | null;
    clientRect?: (() => DOMRect) | null;
}

interface SuggestionKeyDownProps {
    view: unknown; // Prosemirror view, avoiding deep dependency
    event: KeyboardEvent;
    range: Range;
}

interface MentionListProps {
    items: SuggestionItem[];
    command: (props: { id: string; label: string }) => void;
}

const MentionList = React.forwardRef((props: MentionListProps, ref) => {
    const [selectedIndex, setSelectedIndex] = React.useState(0);

    const selectItem = (index: number) => {
        const item = props.items[index];
        if (item) {
            props.command({ id: item.id, label: item.label });
        }
    };

    React.useImperativeHandle(ref, () => ({
        onKeyDown: ({ event }: { event: KeyboardEvent }) => {
            if (event.key === KEYS.ARROW_UP) {
                setSelectedIndex((selectedIndex + props.items.length - 1) % props.items.length);
                return true;
            }
            if (event.key === KEYS.ARROW_DOWN) {
                setSelectedIndex((selectedIndex + 1) % props.items.length);
                return true;
            }
            if (event.key === KEYS.ENTER) {
                selectItem(selectedIndex);
                return true;
            }
            return false;
        },
    }));

    return (
        <div className="bg-white border border-slate-200 rounded-lg shadow-lg overflow-hidden min-w-[150px]">
            {props.items.length ? (
                props.items.map((item, index) => (
                    <button
                        key={item.id}
                        className={`block w-full text-left px-3 py-2 text-sm ${index === selectedIndex ? 'bg-blue-50 text-blue-600' : 'hover:bg-slate-50'
                            }`}
                        onClick={() => selectItem(index)}
                    >
                        {item.label}
                    </button>
                ))
            ) : (
                <div className="px-3 py-2 text-sm text-slate-400">No result</div>
            )}
        </div>
    );
});

export const RichTextEditor: React.FC<RichTextEditorProps> = ({ value, onChange, placeholder, mentionItems = [] }) => {
    const editor = useEditor({
        extensions: [
            StarterKit,
            Mention.configure({
                HTMLAttributes: {
                    class: 'mention text-blue-600 font-medium bg-blue-50 px-1 rounded',
                },
                suggestion: {
                    items: ({ query }) => {
                        return mentionItems
                            .filter(item => item.label.toLowerCase().startsWith(query.toLowerCase()))
                            .slice(0, 5);
                    },
                    render: () => {
                        let component: ReactRenderer;
                        let popup: unknown[];

                        return {
                            onStart: (props: SuggestionProps) => {
                                component = new ReactRenderer(MentionList, {
                                    props,
                                    editor: props.editor,
                                });

                                if (!props.clientRect) return;

                                popup = tippy('body', {
                                    getReferenceClientRect: props.clientRect as () => DOMRect,
                                    appendTo: () => document.body,
                                    content: component.element,
                                    showOnCreate: true,
                                    interactive: true,
                                    trigger: 'manual',
                                    placement: 'bottom-start',
                                }) as unknown as unknown[];
                            },
                            onUpdate: (props: SuggestionProps) => {
                                component.updateProps(props);
                                if (!props.clientRect) return;

                                const tippyInstance = popup?.[0] as { setProps: (props: unknown) => void } | undefined;
                                tippyInstance?.setProps({
                                    getReferenceClientRect: props.clientRect,
                                });
                            },
                            onKeyDown: (props: SuggestionKeyDownProps) => {
                                if (props.event.key === KEYS.ESCAPE) {
                                    const tippyInstance = popup?.[0] as { hide: () => void } | undefined;
                                    tippyInstance?.hide();
                                    return true;
                                }

                                const ref = component.ref as { onKeyDown: (props: SuggestionKeyDownProps) => boolean } | null;
                                return ref?.onKeyDown(props) ?? false;
                            },
                            onExit: () => {
                                const tippyInstance = popup?.[0] as { destroy: () => void } | undefined;
                                tippyInstance?.destroy();
                                component.destroy();
                            },
                        };
                    },
                },
            }),
        ],
        content: value,
        onUpdate: ({ editor }) => {
            onChange(editor.getHTML());
        },
        editorProps: {
            attributes: {
                class: 'prose prose-sm focus:outline-none min-h-[80px] px-3 py-2',
            },
        },
    });

    // Handle external value changes (reset)
    React.useEffect(() => {
        if (editor && value === '') {
            editor.commands.setContent('');
        }
    }, [value, editor]);

    if (!editor) {
        return null;
    }

    return (
        <div className="border border-slate-300 rounded-lg overflow-hidden bg-white focus-within:ring-2 focus-within:ring-blue-500/20 focus-within:border-blue-500 transition-all">
            <MenuBar editor={editor} />
            <div className="bg-white">
                <EditorContent editor={editor} />
            </div>
        </div>
    );
};

// Simple Menu Bar
const MenuBar = ({ editor }: { editor: Editor | null }) => {
    if (!editor) {
        return null;
    }

    const Button = ({ onClick, isActive, children, title }: { onClick: () => void; isActive: boolean; children: React.ReactNode; title: string }) => (
        <button
            onClick={onClick}
            className={`p-1.5 rounded hover:bg-slate-100 ${isActive ? 'bg-slate-200 text-slate-900' : 'text-slate-500'}`}
            title={title}
            type="button"
        >
            {children}
        </button>
    );

    return (
        <div className="flex gap-1 p-2 border-b border-slate-100 bg-slate-50 items-center">
            <Button
                onClick={() => editor.chain().focus().toggleBold().run()}
                isActive={editor.isActive('bold')}
                title="Bold"
            >
                <span className="font-bold">B</span>
            </Button>
            <Button
                onClick={() => editor.chain().focus().toggleItalic().run()}
                isActive={editor.isActive('italic')}
                title="Italic"
            >
                <span className="italic">I</span>
            </Button>
            <Button
                onClick={() => editor.chain().focus().toggleBulletList().run()}
                isActive={editor.isActive('bulletList')}
                title="Bullet List"
            >
                <span>• List</span>
            </Button>
            <Button
                onClick={() => editor.chain().focus().toggleOrderedList().run()}
                isActive={editor.isActive('orderedList')}
                title="Number List"
            >
                <span>1. List</span>
            </Button>
        </div>
    );
};
