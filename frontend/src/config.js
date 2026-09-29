/**
 * Central runtime configuration for the frontend.
 *
 * REACT_APP_API_BASE_URL is a Create React App env var, so it gets
 * inlined into the build at build time (`npm run build`). It controls
 * both the REST API base URL and the WebSocket host, so the frontend
 * can point at a backend that isn't localhost:8080 without any code
 * changes — just a different value at build/deploy time.
 */
export const API_BASE_URL = process.env.REACT_APP_API_BASE_URL || 'http://localhost:8080';
