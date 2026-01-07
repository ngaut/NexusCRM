import { LOGGING_CONFIG } from '../constants/EnvironmentConfig';

enum LogLevel {
    ERROR = 0,
    WARN = 1,
    INFO = 2,
    DEBUG = 3,
    TRACE = 4,
}

const LEVEL_MAP: Record<string, LogLevel> = {
    'error': LogLevel.ERROR,
    'warn': LogLevel.WARN,
    'info': LogLevel.INFO,
    'debug': LogLevel.DEBUG,
    'trace': LogLevel.TRACE,
};

class LoggerService {
    private currentLevel: LogLevel;

    constructor() {
        this.currentLevel = LEVEL_MAP[LOGGING_CONFIG.LOG_LEVEL] ?? LogLevel.INFO;
    }

    private shouldLog(level: LogLevel): boolean {
        return level <= this.currentLevel;
    }

    private formatMessage(level: string, message: string, ...args: any[]): void {
        const timestamp = new Date().toISOString();
        // In dev, shorter format. In prod, maybe JSON? For now, stick to console for browser viewing.
        // We can add remote logging hooks here later.
    }

    public error(message: string, ...args: any[]): void {
        if (this.shouldLog(LogLevel.ERROR)) {
            console.error(`[ERROR] ${message}`, ...args);
        }
    }

    public warn(message: string, ...args: any[]): void {
        if (this.shouldLog(LogLevel.WARN)) {
            console.warn(`[WARN] ${message}`, ...args);
        }
    }

    public info(message: string, ...args: any[]): void {
        if (this.shouldLog(LogLevel.INFO)) {
            console.log(`[INFO] ${message}`, ...args);
        }
    }

    public debug(message: string, ...args: any[]): void {
        if (this.shouldLog(LogLevel.DEBUG)) {
            console.debug(`[DEBUG] ${message}`, ...args);
        }
    }

    public trace(message: string, ...args: any[]): void {
        if (this.shouldLog(LogLevel.TRACE)) {
            console.trace(`[TRACE] ${message}`, ...args);
        }
    }
}

export const Logger = new LoggerService();
