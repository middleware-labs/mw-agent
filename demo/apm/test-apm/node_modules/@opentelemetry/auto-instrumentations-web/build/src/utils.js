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
exports.getWebAutoInstrumentations = void 0;
const api_1 = require("@opentelemetry/api");
const instrumentation_document_load_1 = require("@opentelemetry/instrumentation-document-load");
const instrumentation_fetch_1 = require("@opentelemetry/instrumentation-fetch");
const instrumentation_user_interaction_1 = require("@opentelemetry/instrumentation-user-interaction");
const instrumentation_xml_http_request_1 = require("@opentelemetry/instrumentation-xml-http-request");
const InstrumentationMap = {
    '@opentelemetry/instrumentation-document-load': instrumentation_document_load_1.DocumentLoadInstrumentation,
    '@opentelemetry/instrumentation-fetch': instrumentation_fetch_1.FetchInstrumentation,
    '@opentelemetry/instrumentation-user-interaction': instrumentation_user_interaction_1.UserInteractionInstrumentation,
    '@opentelemetry/instrumentation-xml-http-request': instrumentation_xml_http_request_1.XMLHttpRequestInstrumentation,
};
function getWebAutoInstrumentations(inputConfigs = {}) {
    var _a;
    for (const name of Object.keys(inputConfigs)) {
        if (!Object.prototype.hasOwnProperty.call(InstrumentationMap, name)) {
            api_1.diag.error(`Provided instrumentation name "${name}" not found`);
            continue;
        }
    }
    const instrumentations = [];
    for (const name of Object.keys(InstrumentationMap)) {
        const Instance = InstrumentationMap[name];
        // Defaults are defined by the instrumentation itself
        const userConfig = (_a = inputConfigs[name]) !== null && _a !== void 0 ? _a : {};
        if (userConfig.enabled === false) {
            api_1.diag.debug(`Disabling instrumentation for ${name}`);
            continue;
        }
        try {
            api_1.diag.debug(`Loading instrumentation for ${name}`);
            instrumentations.push(new Instance(userConfig));
        }
        catch (e) {
            api_1.diag.error(e);
        }
    }
    return instrumentations;
}
exports.getWebAutoInstrumentations = getWebAutoInstrumentations;
//# sourceMappingURL=utils.js.map