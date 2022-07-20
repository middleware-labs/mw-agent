import { Context, ContextManager } from '@opentelemetry/api';
/**
 * Stack Context Manager for managing the state in web
 * it doesn't fully support the async calls though
 */
export declare class StackContextManager implements ContextManager {
    /**
     * whether the context manager is enabled or not
     */
    private _enabled;
    /**
     * Keeps the reference to current context
     */
    _currentContext: Context;
    /**
     *
     * @param context
     * @param target Function to be executed within the context
     */
    private _bindFunction;
    /**
     * Returns the active context
     */
    active(): Context;
    /**
     * Binds a the certain context or the active one to the target function and then returns the target
     * @param context A context (span) to be bind to target
     * @param target a function or event emitter. When target or one of its callbacks is called,
     *  the provided context will be used as the active context for the duration of the call.
     */
    bind<T>(context: Context, target: T): T;
    /**
     * Disable the context manager (clears the current context)
     */
    disable(): this;
    /**
     * Enables the context manager and creates a default(root) context
     */
    enable(): this;
    /**
     * Calls the callback function [fn] with the provided [context]. If [context] is undefined then it will use the window.
     * The context will be set as active
     * @param context
     * @param fn Callback function
     * @param thisArg optional receiver to be used for calling fn
     * @param args optional arguments forwarded to fn
     */
    with<A extends unknown[], F extends (...args: A) => ReturnType<F>>(context: Context | null, fn: F, thisArg?: ThisParameterType<F>, ...args: A): ReturnType<F>;
}
//# sourceMappingURL=StackContextManager.d.ts.map