/**
* Copyright 2018 Comcast Cable Communications Management, LLC
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
* http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package config

const (
	defaultLogFile  = ""
	defaultLogLevel = "INFO"

	defaultProxyListenPort    = 9090
	defaultProxyListenAddress = ""

	defaultMetricsListenPort    = 8082
	defaultMetricsListenAddress = ""

	defaultTracerImplemetation    = "opentelemetry"
	defaultExporterImplementation = "noop"

	defaultCacheType   = "memory"
	defaultCacheTypeID = CacheTypeMemory

	defaultTimeseriesTTLSecs     = 21600 * 4 * 30
	defaultFastForwardTTLSecs    = 15
	defaultMaxTTLSecs            = 86400
	defaultMissingToleranceRatio = 0.05
	defaultRevalidationFactor    = 2

	defaultCachePath = "/tmp/trickster"

	defaultRedisClientType = "standard"
	defaultRedisProtocol   = "tcp"
	defaultRedisEndpoint   = "redis:6379"

	defaultBBoltFile   = "trickster.db"
	defaultBBoltBucket = "trickster"

	defaultCacheIndexReap        = 3
	defaultCacheIndexFlush       = 5
	defaultCacheMaxSizeBytes     = 536870912
	defaultMaxSizeBackoffBytes   = 16777216
	defaultMaxSizeObjects        = 0
	defaultMaxSizeBackoffObjects = 100
	defaultMaxObjectSizeBytes    = 524288

	defaultOriginTRF               = 1024
	defaultOriginTEM               = EvictionMethodOldest
	defaultOriginTEMName           = "oldest"
	defaultOriginTimeoutSecs       = 180
	defaultOriginCacheName         = "default"
	defaultOriginNegativeCacheName = "default"
	defaultTracingConfigName       = "default"
	defaultBackfillToleranceSecs   = 0
	defaultKeepAliveTimeoutSecs    = 300
	defaultMaxIdleConns            = 20

	defaultHealthCheckPath  = "-"
	defaultHealthCheckQuery = "-"
	defaultHealthCheckVerb  = "-"

	defaultConfigHandlerPath = "/trickster/config"
	defaultPingHandlerPath   = "/trickster/ping"
)

func defaultCompressableTypes() []string {
	return []string{
		"text/html",
		"text/javascript",
		"text/css",
		"text/plain",
		"text/xml",
		"text/json",
		"application/json",
		"application/javascript",
		"application/xml",
	}
}
