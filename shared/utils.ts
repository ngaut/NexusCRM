/**
 * RegistryBase - Abstract base class for all registries
 * Provides common functionality for registration, validation, and retrieval
 */
export class RegistryBase<T> {
    protected items: Map<string, T> = new Map();
    protected options: {
        validator?: (key: string, value: T) => boolean | string;
        enableEvents?: boolean;
        allowOverwrite?: boolean;
    };

    constructor(options: {
        validator?: (key: string, value: T) => boolean | string;
        enableEvents?: boolean;
        allowOverwrite?: boolean;
    } = {}) {
        this.options = options;
    }

    /**
     * Register an item
     * @param key Unique key for the item
     * @param value The item to register
     */
    public register(key: string, value: T): void {
        if (this.items.has(key) && !this.options.allowOverwrite) {
            console.warn(`Registry: Item with key "${key}" already exists. Overwrite denied.`);
            return;
        }

        if (this.options.validator) {
            const validationResult = this.options.validator(key, value);
            if (validationResult !== true) {
                console.error(`Registry: Validation failed for "${key}": ${validationResult}`);
                return;
            }
        }

        this.items.set(key, value);

        if (this.options.enableEvents) {
            // Event system simplified
        }
    }

    /**
     * Get an item by key
     * @param key The key to look up
     */
    public get(key: string): T | undefined {
        return this.items.get(key);
    }

    /**
     * Check if an item exists
     */
    public has(key: string): boolean {
        return this.items.has(key);
    }

    /**
     * Get all registered keys
     */
    public getKeys(): string[] {
        return Array.from(this.items.keys());
    }

    /**
     * Get all registered values
     */
    public getValues(): T[] {
        return Array.from(this.items.values());
    }

    /**
     * Clear the registry
     */
    public clear(): void {
        this.items.clear();
    }
}
