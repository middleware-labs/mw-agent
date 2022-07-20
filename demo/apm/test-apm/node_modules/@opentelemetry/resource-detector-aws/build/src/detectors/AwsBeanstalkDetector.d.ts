import { Detector, Resource, ResourceDetectionConfig } from '@opentelemetry/resources';
export declare class AwsBeanstalkDetector implements Detector {
    BEANSTALK_CONF_PATH: string;
    private static readFileAsync;
    private static fileAccessAsync;
    constructor();
    detect(_config?: ResourceDetectionConfig): Promise<Resource>;
}
export declare const awsBeanstalkDetector: AwsBeanstalkDetector;
//# sourceMappingURL=AwsBeanstalkDetector.d.ts.map