// Sample JavaScript project for integration testing

const lodash = require('lodash');
const fs = require('fs');
const path = require('path');

// Main application class
class Application {
    constructor(config) {
        this.config = config || {};
        this.initialized = false;
        this.modules = new Map();
    }

    /**
     * Initialize the application with configuration
     * @param {Object} options - Configuration options
     * @returns {Promise<boolean>} Success status
     */
    async initialize(options = {}) {
        try {
            this.config = { ...this.config, ...options };
            
            // Load required modules
            await this.loadModules();
            
            // Setup event listeners
            this.setupEventListeners();
            
            this.initialized = true;
            return true;
        } catch (error) {
            console.error('Application initialization failed:', error);
            return false;
        }
    }

    async loadModules() {
        const moduleDir = path.join(__dirname, 'modules');
        const files = await fs.promises.readdir(moduleDir);
        
        for (const file of files) {
            if (file.endsWith('.js')) {
                const moduleName = path.basename(file, '.js');
                const ModuleClass = require(path.join(moduleDir, file));
                this.modules.set(moduleName, new ModuleClass(this.config));
            }
        }
    }

    setupEventListeners() {
        process.on('SIGINT', () => {
            this.shutdown();
        });

        process.on('uncaughtException', (error) => {
            console.error('Uncaught Exception:', error);
            this.shutdown();
        });
    }

    getModule(name) {
        return this.modules.get(name);
    }

    async shutdown() {
        console.log('Shutting down application...');
        
        for (const [name, module] of this.modules) {
            if (typeof module.cleanup === 'function') {
                await module.cleanup();
            }
        }
        
        process.exit(0);
    }
}

// Utility functions
const utils = {
    formatDate: (date) => {
        return new Intl.DateTimeFormat('en-US').format(date);
    },

    validateEmail: (email) => {
        const regex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return regex.test(email);
    },

    deepClone: (obj) => {
        return JSON.parse(JSON.stringify(obj));
    },

    debounce: (func, wait) => {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }
};

// Data processing functions
function processUserData(users) {
    return users
        .filter(user => user.active && utils.validateEmail(user.email))
        .map(user => ({
            ...user,
            displayName: `${user.firstName} ${user.lastName}`,
            joinDate: utils.formatDate(new Date(user.createdAt))
        }))
        .sort((a, b) => a.lastName.localeCompare(b.lastName));
}

function calculateStatistics(data) {
    if (!Array.isArray(data) || data.length === 0) {
        return { mean: 0, median: 0, mode: 0, range: 0 };
    }

    const sorted = [...data].sort((a, b) => a - b);
    const sum = data.reduce((acc, val) => acc + val, 0);
    const mean = sum / data.length;
    
    const median = sorted.length % 2 === 0
        ? (sorted[sorted.length / 2 - 1] + sorted[sorted.length / 2]) / 2
        : sorted[Math.floor(sorted.length / 2)];

    const frequency = {};
    let maxCount = 0;
    let mode = data[0];

    for (const num of data) {
        frequency[num] = (frequency[num] || 0) + 1;
        if (frequency[num] > maxCount) {
            maxCount = frequency[num];
            mode = num;
        }
    }

    const range = Math.max(...data) - Math.min(...data);

    return { mean, median, mode, range };
}

// API interaction helpers
const api = {
    baseURL: process.env.API_BASE_URL || 'http://localhost:3000/api',

    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers
            },
            ...options
        };

        try {
            const response = await fetch(url, config);
            
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            return await response.json();
        } catch (error) {
            console.error('API request failed:', error);
            throw error;
        }
    },

    get: (endpoint, options = {}) => api.request(endpoint, { method: 'GET', ...options }),
    post: (endpoint, data, options = {}) => api.request(endpoint, { 
        method: 'POST', 
        body: JSON.stringify(data), 
        ...options 
    }),
    put: (endpoint, data, options = {}) => api.request(endpoint, { 
        method: 'PUT', 
        body: JSON.stringify(data), 
        ...options 
    }),
    delete: (endpoint, options = {}) => api.request(endpoint, { method: 'DELETE', ...options })
};

// Export all functionality
module.exports = {
    Application,
    utils,
    processUserData,
    calculateStatistics,
    api
};