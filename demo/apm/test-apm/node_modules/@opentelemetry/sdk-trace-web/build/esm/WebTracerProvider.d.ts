import { BasicTracerProvider, SDKRegistrationConfig, TracerConfig } from '@opentelemetry/sdk-trace-base';
/**
 * WebTracerConfig provides an interface for configuring a Web Tracer.
 */
export declare type WebTracerConfig = TracerConfig;
/**
 * This class represents a web tracer with {@link StackContextManager}
 */
export declare class WebTracerProvider extends BasicTracerProvider {
    /**
     * Constructs a new Tracer instance.
     * @param config Web Tracer config
     */
    constructor(config?: WebTracerConfig);
    /**
     * Register this TracerProvider for use with the OpenTelemetry API.
     * Undefined values may be replaced with defaults, and
     * null values will be skipped.
     *
     * @param config Configuration object for SDK registration
     */
    register(config?: SDKRegistrationConfig): void;
}
//# sourceMappingURL=WebTracerProvider.d.ts.map