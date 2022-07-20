import * as api from '@opentelemetry/api';
/**
 * Interface used to provide information to finish span on fetch response
 */
export interface FetchResponse {
    status: number;
    statusText?: string;
    url: string;
}
/**
 * Interface used to provide information to finish span on fetch error
 */
export interface FetchError {
    status?: number;
    message: string;
}
/**
 * Interface used to keep information about span between creating and
 * ending span
 */
export interface SpanData {
    entries: PerformanceResourceTiming[];
    observer?: PerformanceObserver;
    spanUrl: string;
    startTime: api.HrTime;
}
//# sourceMappingURL=types.d.ts.map