// Sample TypeScript project for integration testing

import { EventEmitter } from 'events';
import { promises as fs } from 'fs';
import { join } from 'path';

// Type definitions
interface User {
    id: number;
    firstName: string;
    lastName: string;
    email: string;
    active: boolean;
    createdAt: string;
    role: UserRole;
}

interface ProcessedUser extends User {
    displayName: string;
    joinDate: string;
}

type UserRole = 'admin' | 'user' | 'moderator';

interface AppConfig {
    apiBaseURL?: string;
    debugMode?: boolean;
    maxConnections?: number;
    timeout?: number;
}

interface Statistics {
    mean: number;
    median: number;
    mode: number;
    range: number;
    standardDeviation: number;
}

// Generic interfaces
interface Repository<T> {
    findById(id: number): Promise<T | null>;
    findAll(): Promise<T[]>;
    create(data: Omit<T, 'id'>): Promise<T>;
    update(id: number, data: Partial<T>): Promise<T | null>;
    delete(id: number): Promise<boolean>;
}

interface Cacheable {
    cacheKey: string;
    expiresAt: Date;
}

// Main application class with generics
class TypedApplication<TConfig extends AppConfig> extends EventEmitter {
    private readonly config: TConfig;
    private initialized: boolean = false;
    private readonly modules: Map<string, any> = new Map();
    private readonly cache: Map<string, Cacheable> = new Map();

    constructor(config: TConfig) {
        super();
        this.config = { ...config };
    }

    async initialize(additionalConfig?: Partial<TConfig>): Promise<boolean> {
        try {
            if (additionalConfig) {
                Object.assign(this.config, additionalConfig);
            }

            await this.loadModules();
            this.setupEventHandlers();

            this.initialized = true;
            this.emit('initialized');
            
            return true;
        } catch (error) {
            this.emit('error', error);
            return false;
        }
    }

    private async loadModules(): Promise<void> {
        const moduleDir = join(__dirname, 'modules');
        
        try {
            const files = await fs.readdir(moduleDir);
            
            for (const file of files) {
                if (file.endsWith('.ts') || file.endsWith('.js')) {
                    const moduleName = file.replace(/\.(ts|js)$/, '');
                    const ModuleClass = await import(join(moduleDir, file));
                    
                    this.modules.set(moduleName, new ModuleClass.default(this.config));
                }
            }
        } catch (error) {
            throw new Error(`Failed to load modules: ${error}`);
        }
    }

    private setupEventHandlers(): void {
        process.on('SIGINT', () => this.shutdown());
        process.on('SIGTERM', () => this.shutdown());
        
        this.on('error', (error) => {
            console.error('Application error:', error);
        });
    }

    getModule<T = any>(name: string): T | undefined {
        return this.modules.get(name) as T;
    }

    getCachedData<T extends Cacheable>(key: string): T | null {
        const cached = this.cache.get(key);
        
        if (!cached || cached.expiresAt < new Date()) {
            this.cache.delete(key);
            return null;
        }
        
        return cached as T;
    }

    setCachedData<T extends Cacheable>(key: string, data: T, ttl: number = 300000): void {
        data.expiresAt = new Date(Date.now() + ttl);
        this.cache.set(key, data);
    }

    async shutdown(): Promise<void> {
        console.log('Shutting down TypeScript application...');
        
        for (const [name, module] of this.modules) {
            if (module && typeof module.cleanup === 'function') {
                await module.cleanup();
            }
        }
        
        this.emit('shutdown');
        process.exit(0);
    }
}

// Service classes with dependency injection
abstract class BaseService {
    protected readonly name: string;
    protected initialized: boolean = false;

    constructor(name: string) {
        this.name = name;
    }

    abstract initialize(): Promise<void>;
    abstract cleanup(): Promise<void>;
}

class UserService extends BaseService implements Repository<User> {
    private users: User[] = [];

    constructor() {
        super('UserService');
    }

    async initialize(): Promise<void> {
        // Mock initialization
        this.initialized = true;
    }

    async findById(id: number): Promise<User | null> {
        return this.users.find(user => user.id === id) || null;
    }

    async findAll(): Promise<User[]> {
        return [...this.users];
    }

    async create(data: Omit<User, 'id'>): Promise<User> {
        const user: User = {
            id: Date.now(),
            ...data
        };
        
        this.users.push(user);
        return user;
    }

    async update(id: number, data: Partial<User>): Promise<User | null> {
        const userIndex = this.users.findIndex(user => user.id === id);
        
        if (userIndex === -1) {
            return null;
        }
        
        this.users[userIndex] = { ...this.users[userIndex], ...data };
        return this.users[userIndex];
    }

    async delete(id: number): Promise<boolean> {
        const initialLength = this.users.length;
        this.users = this.users.filter(user => user.id !== id);
        
        return this.users.length < initialLength;
    }

    async cleanup(): Promise<void> {
        this.users = [];
        this.initialized = false;
    }
}

// Utility functions with proper typing
class MathUtils {
    static calculateStatistics(data: number[]): Statistics {
        if (data.length === 0) {
            return { mean: 0, median: 0, mode: 0, range: 0, standardDeviation: 0 };
        }

        const sorted = [...data].sort((a, b) => a - b);
        const sum = data.reduce((acc, val) => acc + val, 0);
        const mean = sum / data.length;

        // Calculate median
        const median = sorted.length % 2 === 0
            ? (sorted[sorted.length / 2 - 1] + sorted[sorted.length / 2]) / 2
            : sorted[Math.floor(sorted.length / 2)];

        // Calculate mode
        const frequency: Record<number, number> = {};
        let maxCount = 0;
        let mode = data[0];

        for (const num of data) {
            frequency[num] = (frequency[num] || 0) + 1;
            if (frequency[num] > maxCount) {
                maxCount = frequency[num];
                mode = num;
            }
        }

        // Calculate range
        const range = Math.max(...data) - Math.min(...data);

        // Calculate standard deviation
        const variance = data.reduce((acc, val) => acc + Math.pow(val - mean, 2), 0) / data.length;
        const standardDeviation = Math.sqrt(variance);

        return { mean, median, mode, range, standardDeviation };
    }

    static clamp(value: number, min: number, max: number): number {
        return Math.min(Math.max(value, min), max);
    }

    static lerp(start: number, end: number, factor: number): number {
        return start + (end - start) * this.clamp(factor, 0, 1);
    }
}

// Data processing with generics
class DataProcessor<T> {
    private data: T[] = [];

    constructor(initialData: T[] = []) {
        this.data = [...initialData];
    }

    add(item: T): void {
        this.data.push(item);
    }

    filter(predicate: (item: T) => boolean): T[] {
        return this.data.filter(predicate);
    }

    map<U>(transform: (item: T) => U): U[] {
        return this.data.map(transform);
    }

    reduce<U>(reducer: (acc: U, current: T) => U, initialValue: U): U {
        return this.data.reduce(reducer, initialValue);
    }

    sort(compareFn?: (a: T, b: T) => number): T[] {
        return [...this.data].sort(compareFn);
    }

    get length(): number {
        return this.data.length;
    }

    get isEmpty(): boolean {
        return this.data.length === 0;
    }
}

// API client with proper error handling
class APIClient {
    private readonly baseURL: string;
    private readonly timeout: number;

    constructor(baseURL: string, timeout: number = 5000) {
        this.baseURL = baseURL;
        this.timeout = timeout;
    }

    async request<T>(
        endpoint: string, 
        options: RequestInit = {}
    ): Promise<T> {
        const url = `${this.baseURL}${endpoint}`;
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), this.timeout);

        try {
            const response = await fetch(url, {
                ...options,
                signal: controller.signal,
                headers: {
                    'Content-Type': 'application/json',
                    ...options.headers
                }
            });

            clearTimeout(timeoutId);

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            return await response.json() as T;
        } catch (error) {
            clearTimeout(timeoutId);
            
            if (error.name === 'AbortError') {
                throw new Error('Request timeout');
            }
            
            throw error;
        }
    }

    async get<T>(endpoint: string): Promise<T> {
        return this.request<T>(endpoint, { method: 'GET' });
    }

    async post<T>(endpoint: string, data: any): Promise<T> {
        return this.request<T>(endpoint, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    }

    async put<T>(endpoint: string, data: any): Promise<T> {
        return this.request<T>(endpoint, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    }

    async delete<T>(endpoint: string): Promise<T> {
        return this.request<T>(endpoint, { method: 'DELETE' });
    }
}

// Export all functionality
export {
    User,
    ProcessedUser,
    UserRole,
    AppConfig,
    Statistics,
    Repository,
    Cacheable,
    TypedApplication,
    BaseService,
    UserService,
    MathUtils,
    DataProcessor,
    APIClient
};

export default TypedApplication;