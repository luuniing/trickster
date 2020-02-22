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

package irondb

import (
	"sync"
	"time"

	"github.com/Comcast/trickster/internal/timeseries"
)

// DF4SeriesEnvelope values represent DF4 format time series data from the
// IRONdb API.
type DF4SeriesEnvelope struct {
	Data         [][]interface{}          `json:"data"`
	Meta         []map[string]interface{} `json:"meta,omitempty"`
	Ver          string                   `json:"version,omitempty"`
	Head         DF4Info                  `json:"head"`
	StepDuration time.Duration            `json:"step,omitempty"`
	ExtentList   timeseries.ExtentList    `json:"extents,omitempty"`
}

// DF4Info values contain information about the timestamps of the data elements
// in DF4 data series.
type DF4Info struct {
	Count  int64 `json:"count"`
	Start  int64 `json:"start"`
	Period int64 `json:"period"`
}

// Step returns the step for the Timeseries.
func (se *DF4SeriesEnvelope) Step() time.Duration {
	return se.StepDuration
}

// SetStep sets the step for the Timeseries.
func (se *DF4SeriesEnvelope) SetStep(step time.Duration) {
	se.StepDuration = step
}

// Extents returns the Timeseries's extent list.
func (se *DF4SeriesEnvelope) Extents() timeseries.ExtentList {
	return se.ExtentList
}

// SetExtents overwrites a Timeseries's known extents with the provided extent
// list.
func (se *DF4SeriesEnvelope) SetExtents(extents timeseries.ExtentList) {
	se.ExtentList = extents
}

// SeriesCount returns the number of individual series in the Timeseries value.
func (se *DF4SeriesEnvelope) SeriesCount() int {
	return len(se.Data)
}

// ValueCount returns the count of all data values across all Series in the
// Timeseries value.
func (se *DF4SeriesEnvelope) ValueCount() int {
	n := 0
	for _, v := range se.Data {
		n += len(v)
	}

	return n
}

// TimestampCount returns the number of unique timestamps across the timeseries.
func (se *DF4SeriesEnvelope) TimestampCount() int {
	return int(se.Head.Count)
}

type metricData struct {
	name string
	meta map[string]interface{}
	data map[int64]interface{}
}

func (se *DF4SeriesEnvelope) SyncExtentFromSamples() {
}

// Merge merges the provided Timeseries list into the base Timeseries (in the
// order provided) and optionally sorts the merged Timeseries.
func (se *DF4SeriesEnvelope) Merge(sort bool,
	collection ...timeseries.Timeseries) {
	for _, ts := range collection {
		if ts != nil && ts.Step() == se.Step() {
			if se2, ok := ts.(*DF4SeriesEnvelope); ok {
				// Build new data series for each metric.
				metrics := map[string]*metricData{}
				for i, mv := range se.Meta {
					if name, ok := mv["label"].(string); ok {
						metrics[name] = &metricData{
							name: name,
							meta: mv,
							data: map[int64]interface{}{},
						}

						for j, dv := range se.Data[i] {
							ts := se.Head.Start + (int64(j) * se.Head.Period)
							metrics[name].data[ts] = dv
						}
					}
				}

				// Merge in the data from the merging series.
				for i, mv := range se2.Meta {
					if name, ok := mv["label"].(string); ok {
						md, ok := metrics[name]
						if !ok {
							metrics[name] = &metricData{
								name: name,
								meta: mv,
								data: map[int64]interface{}{},
							}

							md = metrics[name]
						}

						for j, dv := range se2.Data[i] {
							ts := se2.Head.Start +
								(int64(j) * se2.Head.Period)
							md.data[ts] = dv
						}
					}
				}

				// Calculate the new range of data points.
				min := se.Head.Start
				if se2.Head.Start < se.Head.Start {
					min = se2.Head.Start
				}

				max := se.Head.Start + ((se.Head.Count - 1) * se.Head.Period)
				max2 := se2.Head.Start + ((se2.Head.Count - 1) * se2.Head.Period)
				if max2 > max {
					max = max2
				}

				// Merge the new data series.
				newData := [][]interface{}{}
				newMeta := []map[string]interface{}{}
				newHead := DF4Info{
					Count:  (max-min)/se.Head.Period + 1,
					Start:  min,
					Period: se.Head.Period,
				}

				for _, m := range metrics {
					newMeta = append(newMeta, m.meta)
					d := []interface{}{}
					for i := int64(0); i < newHead.Count; i++ {
						ts := newHead.Start + (i * newHead.Period)
						d = append(d, m.data[ts])
					}

					newData = append(newData, d)
				}

				se.Data = newData
				se.Meta = newMeta
				se.Head = newHead
				se.ExtentList = append(se.ExtentList, se2.ExtentList...)
			}
		}
	}

	se.ExtentList = se.ExtentList.Compress(se.StepDuration)
	if sort {
		se.Sort()
	}
}

// Clone returns a perfect copy of the base Timeseries.
func (se *DF4SeriesEnvelope) Clone() timeseries.Timeseries {
	b := &DF4SeriesEnvelope{
		Data: make([][]interface{}, len(se.Data)),
		Meta: make([]map[string]interface{}, len(se.Meta)),
		Ver:  se.Ver,
		Head: DF4Info{
			Count:  se.Head.Count,
			Start:  se.Head.Start,
			Period: se.Head.Period,
		},
		StepDuration: se.StepDuration,
		ExtentList:   se.ExtentList.Clone(),
	}

	for i, v := range se.Data {
		b.Data[i] = make([]interface{}, len(v))
		copy(b.Data[i], v)
	}

	for i, v := range se.Meta {
		b.Meta[i] = make(map[string]interface{}, len(se.Meta[i]))
		for k, mv := range v {
			b.Meta[i][k] = mv
		}
	}

	return b
}

// CropToRange crops down a Timeseries value to the provided Extent.
// Crop assumes the base Timeseries is already sorted, and will corrupt an
// unsorted Timeseries.
func (se *DF4SeriesEnvelope) CropToRange(e timeseries.Extent) {
	// Align crop extents with step period.
	e.Start = time.Unix(e.Start.Unix()-(e.Start.Unix()%se.Head.Period), 0)
	e.End = time.Unix(e.End.Unix()-(e.End.Unix()%se.Head.Period), 0)

	// If the Timeseries has no extents, or the extent of the series is entirely
	// outside the extent of the crop range, return empty set and bail.
	if len(se.ExtentList) < 1 || se.ExtentList.OutsideOf(e) {
		se.Data = [][]interface{}{}
		se.Meta = []map[string]interface{}{}
		se.Head.Start = e.Start.Unix()
		se.Head.Count = 0
		se.ExtentList = timeseries.ExtentList{}
		return
	}

	// Create a map of the time series data.
	metrics := map[string]metricData{}
	for i, mv := range se.Meta {
		if name, ok := mv["label"].(string); ok {
			metrics[name] = metricData{
				name: name,
				meta: mv,
				data: map[int64]interface{}{},
			}

			for j, dv := range se.Data[i] {
				ts := se.Head.Start + (int64(j) * se.Head.Period)
				if ts >= e.Start.Unix() && ts <= e.End.Unix() {
					metrics[name].data[ts] = dv
				}
			}
		}
	}

	// Replace with the cropped data series.
	newData := [][]interface{}{}
	newMeta := []map[string]interface{}{}
	newHead := DF4Info{
		Count:  (e.End.Unix() - e.Start.Unix()) / se.Head.Period,
		Start:  e.Start.Unix(),
		Period: se.Head.Period,
	}

	for _, m := range metrics {
		newMeta = append(newMeta, m.meta)
		d := []interface{}{}
		for i := int64(0); i < newHead.Count; i++ {
			ts := newHead.Start + (i * newHead.Period)
			d = append(d, m.data[ts])
		}

		newData = append(newData, d)
	}

	se.Data = newData
	se.Meta = newMeta
	se.Head = newHead
	se.ExtentList = se.ExtentList.Crop(e)
}

// CropToSize reduces the number of elements in the Timeseries to the provided
// count, by evicting elements using a least-recently-used methodology. Any
// timestamps newer than the provided time are removed before sizing, in order
// to support backfill tolerance. The provided extent will be marked as used
// during crop.
func (se *DF4SeriesEnvelope) CropToSize(sz int, t time.Time,
	lur timeseries.Extent) {
	// The Series has no extents, so no need to do anything.
	if len(se.ExtentList) < 1 {
		se.Data = [][]interface{}{}
		se.Meta = []map[string]interface{}{}
		se.Head.Start = 0
		se.Head.Count = 0
		se.ExtentList = timeseries.ExtentList{}
		return
	}

	// Crop to the Backfill Tolerance Value if needed.
	if se.ExtentList[len(se.ExtentList)-1].End.After(t) {
		se.CropToRange(timeseries.Extent{Start: se.ExtentList[0].Start, End: t})
	}

	tc := se.TimestampCount()
	if len(se.Data) == 0 || tc <= sz {
		return
	}

	rc := tc - sz // removal count
	newData := [][]interface{}{}
	for _, data := range se.Data {
		newData = append(newData, data[rc:])
	}

	se.Head.Start += int64(rc) * se.Head.Period
	se.Head.Count -= int64(rc)
	se.Data = newData
	se.ExtentList = timeseries.ExtentList{timeseries.Extent{
		Start: time.Unix(se.Head.Start, 0),
		End:   time.Unix(se.Head.Start+((se.Head.Count-1)*se.Head.Period), 0),
	}}
}

// Sort sorts all data in the Timeseries chronologically by their timestamp.
func (se *DF4SeriesEnvelope) Sort() {
	// DF4SeriesEnvelope is sorted by definition.
}

// Size returns the approximate memory utilization in bytes of the timeseries
func (se *DF4SeriesEnvelope) Size() int {

	// TODO this implementation is a rough approximation to ensure we conform to the
	// interface specification, it requires refinement in order to be in the ballpark

	c := 24 + len(se.Ver) // accounts for head + ver
	wg := sync.WaitGroup{}
	mtx := sync.Mutex{}
	for i := range se.Data {
		wg.Add(1)
		go func(s []interface{}) {
			mtx.Lock()
			c += (len(s) * 16)
			mtx.Unlock()
			wg.Done()
		}(se.Data[i])
	}
	wg.Wait()
	return c
}
