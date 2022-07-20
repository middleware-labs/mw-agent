import { Detector, Resource, ResourceDetectionConfig } from '@opentelemetry/resources';
/**
 * The AwsEcsDetector can be used to detect if a process is running in AWS
 * ECS and return a {@link Resource} populated with data about the ECS
 * plugins of AWS X-Ray. Returns an empty Resource if detection fails.
 */
export declare class AwsEcsDetector implements Detector {
    readonly CONTAINER_ID_LENGTH = 64;
    readonly DEFAULT_CGROUP_PATH = "/proc/self/cgroup";
    private static readFileAsync;
    detect(_config?: ResourceDetectionConfig): Promise<Resource>;
    /**
     * Read container ID from cgroup file
     * In ECS, even if we fail to find target file
     * or target file does not contain container ID
     * we do not throw an error but throw warning message
     * and then return null string
     */
    private _getContainerId;
}
export declare const awsEcsDetector: AwsEcsDetector;
//# sourceMappingURL=AwsEcsDetector.d.ts.map