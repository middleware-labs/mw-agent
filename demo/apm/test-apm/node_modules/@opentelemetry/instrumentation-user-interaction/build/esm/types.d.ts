import { HrTime, Span } from '@opentelemetry/api';
import { InstrumentationConfig } from '@opentelemetry/instrumentation';
export declare type EventName = keyof HTMLElementEventMap;
export declare type ShouldPreventSpanCreation = (eventType: EventName, element: HTMLElement, span: Span) => boolean | void;
export interface UserInteractionInstrumentationConfig extends InstrumentationConfig {
    /**
     * List of events to instrument (like 'mousedown', 'touchend', 'play' etc).
     * By default only 'click' event is instrumented.
     */
    eventNames?: EventName[];
    /**
     * Callback function called each time new span is being created.
     * Return `true` to prevent span recording.
     * You can also use this handler to enhance created span with extra attributes.
     */
    shouldPreventSpanCreation?: ShouldPreventSpanCreation;
}
/**
 * Async Zone task
 */
export declare type AsyncTask = Task & {
    eventName: EventName;
    target: EventTarget;
    _zone: Zone;
};
/**
 *  Type for patching Zone RunTask function
 */
export declare type RunTaskFunction = (task: AsyncTask, applyThis?: any, applyArgs?: any) => Zone;
/**
 * interface to store information in weak map per span
 */
export interface SpanData {
    hrTimeLastTimeout?: HrTime;
    taskCount: number;
}
/**
 * interface to be able to check Zone presence on window
 */
export interface WindowWithZone {
    Zone: ZoneTypeWithPrototype;
}
/**
 * interface to be able to use prototype in Zone
 */
interface ZonePrototype {
    prototype: any;
}
/**
 * type to be  able to use prototype on Zone
 */
export declare type ZoneTypeWithPrototype = ZonePrototype & Zone;
export {};
//# sourceMappingURL=types.d.ts.map