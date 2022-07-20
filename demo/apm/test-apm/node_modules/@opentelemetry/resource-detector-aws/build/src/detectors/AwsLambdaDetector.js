"use strict";
/*
 * Copyright The OpenTelemetry Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.awsLambdaDetector = exports.AwsLambdaDetector = void 0;
const resources_1 = require("@opentelemetry/resources");
const semantic_conventions_1 = require("@opentelemetry/semantic-conventions");
/**
 * The AwsLambdaDetector can be used to detect if a process is running in AWS Lambda
 * and return a {@link Resource} populated with data about the environment.
 * Returns an empty Resource if detection fails.
 */
class AwsLambdaDetector {
    async detect(_config) {
        const functionName = process.env.AWS_LAMBDA_FUNCTION_NAME;
        if (!functionName) {
            return resources_1.Resource.empty();
        }
        const functionVersion = process.env.AWS_LAMBDA_FUNCTION_VERSION;
        const region = process.env.AWS_REGION;
        const attributes = {
            [semantic_conventions_1.SemanticResourceAttributes.CLOUD_PROVIDER]: String(semantic_conventions_1.CloudProviderValues.AWS),
            [semantic_conventions_1.SemanticResourceAttributes.CLOUD_PLATFORM]: String(semantic_conventions_1.CloudPlatformValues.AWS_LAMBDA),
        };
        if (region) {
            attributes[semantic_conventions_1.SemanticResourceAttributes.CLOUD_REGION] = region;
        }
        if (functionName) {
            attributes[semantic_conventions_1.SemanticResourceAttributes.FAAS_NAME] = functionName;
        }
        if (functionVersion) {
            attributes[semantic_conventions_1.SemanticResourceAttributes.FAAS_VERSION] = functionVersion;
        }
        return new resources_1.Resource(attributes);
    }
}
exports.AwsLambdaDetector = AwsLambdaDetector;
exports.awsLambdaDetector = new AwsLambdaDetector();
//# sourceMappingURL=AwsLambdaDetector.js.map