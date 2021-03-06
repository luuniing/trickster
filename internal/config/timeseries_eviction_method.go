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

import "strconv"

// TimeseriesEvictionMethod enumerates the methodologies for maintaining time series cache data
type TimeseriesEvictionMethod int

const (
	// EvictionMethodOldest indicates that a time series cache object only holds values newer than an explicit date,
	// called the Oldest Cacheable Timestamp, which is calculated with this formula on each request:
	// time.Now().Add(-(config.ValueRetentionFactor * query.Step))
	// This policy is the more performant methodology, because out-of-cache-range determination does not require querying
	// the cache; thus the cache is only accessed for requests that are pre-determined to be cacheable
	EvictionMethodOldest = TimeseriesEvictionMethod(iota)
	// EvictionMethodLRU indicates a that a time series cache object hold up to ValueRetentionFactor number of
	// unique timestamps, removing the least-recently-used timestamps as necessary to to remain at the ValueRetentionFactor
	// This policy is the more compute-intensive, since we must maintain an LRU on each timestamp in each cache object,
	// and retrieve the object from cache on each request
	EvictionMethodLRU
)

var timeseriesEvictionMethodNames = map[string]TimeseriesEvictionMethod{
	"oldest": EvictionMethodOldest,
	"lru":    EvictionMethodLRU,
}

var timeseriesEvictionMethodValues = map[TimeseriesEvictionMethod]string{
	EvictionMethodOldest: "oldest",
	EvictionMethodLRU:    "lru",
}

func (t TimeseriesEvictionMethod) String() string {
	if v, ok := timeseriesEvictionMethodValues[t]; ok {
		return v
	}
	return strconv.Itoa(int(t))
}
