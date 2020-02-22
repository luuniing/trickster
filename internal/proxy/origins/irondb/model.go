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
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Comcast/trickster/internal/timeseries"
)

// SeriesEnvelope values represent a time series data response from the
// IRONdb API.
type SeriesEnvelope struct {
	Data         DataPoints            `json:"data"`
	ExtentList   timeseries.ExtentList `json:"extents,omitempty"`
	StepDuration time.Duration         `json:"step,omitempty"`
}

// MarshalJSON encodes a series envelope value into a JSON byte slice.
func (se *SeriesEnvelope) MarshalJSON() ([]byte, error) {
	if se.StepDuration == 0 && len(se.ExtentList) == 0 {
		// Special case for when returning data to the caller.
		return json.Marshal(se.Data)
	}

	se2 := struct {
		Data         DataPoints            `json:"data"`
		ExtentList   timeseries.ExtentList `json:"extents,omitempty"`
		StepDuration string                `json:"step,omitempty"`
	}{
		Data:       se.Data,
		ExtentList: se.ExtentList,
	}

	if se.StepDuration != 0 {
		se2.StepDuration = se.StepDuration.String()
	}

	return json.Marshal(se2)
}

// UnmarshalJSON decodes a JSON byte slice into this data point value.
func (se *SeriesEnvelope) UnmarshalJSON(b []byte) error {
	if strings.Contains(string(b), `"data"`) &&
		(strings.Contains(string(b), `"extents"`) ||
			strings.Contains(string(b), `"step"`)) {
		var se2 struct {
			Data         DataPoints            `json:"data"`
			ExtentList   timeseries.ExtentList `json:"extents,omitempty"`
			StepDuration string                `json:"step,omitempty"`
		}

		if err := json.Unmarshal(b, &se2); err != nil {
			return err
		}

		se.Data = se2.Data
		se.ExtentList = se2.ExtentList
		d, err := time.ParseDuration(se2.StepDuration)
		if err != nil {
			return err
		}

		se.StepDuration = d
		return err
	}

	err := json.Unmarshal(b, &se.Data)
	return err
}

// DataPoint values represent a single data element of a time series data
// response from the IRONdb API.
type DataPoint struct {
	Time  time.Time
	Step  uint32
	Value interface{}
}

// MarshalJSON encodes a data point value into a JSON byte slice.
func (dp *DataPoint) MarshalJSON() ([]byte, error) {
	v := []interface{}{}
	tn := float64(0)
	fv, err := strconv.ParseFloat(formatTimestamp(dp.Time, true), 64)
	if err == nil {
		tn = float64(fv)
	}

	v = append(v, tn)
	if dp.Step != 0 {
		v = append(v, dp.Step)
	}

	v = append(v, dp.Value)
	return json.Marshal(v)
}

// UnmarshalJSON decodes a JSON byte slice into this data point value.
func (dp *DataPoint) UnmarshalJSON(b []byte) error {
	v := []interface{}{}
	err := json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	if len(v) < 2 {
		return fmt.Errorf("unable to unmarshal IRONdb data point: %s",
			string(b))
	}

	if fv, ok := v[0].(float64); ok {
		tv, err := parseTimestamp(strconv.FormatFloat(fv, 'f', 3, 64))
		if err != nil {
			return err
		}

		dp.Time = tv
	}

	if fv, ok := v[1].(float64); ok && len(v) > 2 {
		dp.Step = uint32(fv)
		dp.Value = v[2]
		return nil
	}

	dp.Value = v[1]
	return nil
}

// DataPoints values represent sortable slices of data point values.
type DataPoints []DataPoint

// Len returns the length of an array of Prometheus model.Times
func (dps DataPoints) Len() int {
	return len(dps)
}

// Less returns true if the value at index i comes before the value at index j.
func (dps DataPoints) Less(i, j int) bool {
	return dps[i].Time.Before(dps[j].Time)
}

// Swap modifies a slice of data tuples by swapping the values in indexes
// i and j.
func (dps DataPoints) Swap(i, j int) {
	dps[i], dps[j] = dps[j], dps[i]
}

// Step returns the step for the Timeseries.
func (se *SeriesEnvelope) Step() time.Duration {
	return se.StepDuration
}

// SetStep sets the step for the Timeseries.
func (se *SeriesEnvelope) SetStep(step time.Duration) {
	se.StepDuration = step
}

// SetExtents overwrites a Timeseries's known extents with the provided extent
// list.
func (se *SeriesEnvelope) SetExtents(extents timeseries.ExtentList) {
	se.ExtentList = extents
}

// Extents returns the Timeseries's extent list.
func (se *SeriesEnvelope) Extents() timeseries.ExtentList {
	return se.ExtentList
}

// SeriesCount returns the number of individual series in the Timeseries value.
func (se *SeriesEnvelope) SeriesCount() int {
	return 1
}

// ValueCount returns the count of all data values across all Series in the
// Timeseries value.
func (se *SeriesEnvelope) ValueCount() int {
	return len(se.Data)
}

// TimestampCount returns the number of unique timestamps across the timeseries.
func (se *SeriesEnvelope) TimestampCount() int {
	ts := map[int64]struct{}{}
	for _, dp := range se.Data {
		ts[dp.Time.Unix()] = struct{}{}
	}

	return len(ts)
}

func (se *SeriesEnvelope) SyncExtentFromSamples() {
}

// Merge merges the provided Timeseries list into the base Timeseries (in the
// order provided) and optionally sorts the merged Timeseries.
func (se *SeriesEnvelope) Merge(sort bool,
	collection ...timeseries.Timeseries) {
	for _, ts := range collection {
		if ts != nil {
			if se2, ok := ts.(*SeriesEnvelope); ok {
				se.Data = append(se.Data, se2.Data...)
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
func (se *SeriesEnvelope) Clone() timeseries.Timeseries {
	b := &SeriesEnvelope{
		Data:         make([]DataPoint, len(se.Data)),
		StepDuration: se.StepDuration,
		ExtentList:   make(timeseries.ExtentList, 0, len(se.ExtentList)),
	}

	copy(b.ExtentList, se.ExtentList)
	if len(se.Data) > 0 {
		b.Data = make(DataPoints, len(se.Data))
		copy(b.Data, se.Data)
	}

	return b
}

// CropToRange crops down a Timeseries value to the provided Extent.
// Crop assumes the base Timeseries is already sorted, and will corrupt an
// unsorted Timeseries.
func (se *SeriesEnvelope) CropToRange(e timeseries.Extent) {
	newData := DataPoints{}
	for _, dv := range se.Data {
		if (dv.Time.After(e.Start) || dv.Time.Equal(e.Start)) &&
			(dv.Time.Before(e.End) || dv.Time.Equal(e.End)) {
			newData = append(newData, dv)
		}
	}

	se.Data = newData
	se.ExtentList = se.ExtentList.Crop(e)
}

// CropToSize reduces the number of elements in the Timeseries to the provided
// count, by evicting elements using a least-recently-used methodology. Any
// timestamps newer than the provided time are removed before sizing, in order
// to support backfill tolerance. The provided extent will be marked as used
// during crop.
func (se *SeriesEnvelope) CropToSize(sz int, t time.Time,
	lur timeseries.Extent) {
	// The Series has no extents, so no need to do anything.
	if len(se.ExtentList) < 1 {
		se.Data = DataPoints{}
		se.ExtentList = timeseries.ExtentList{}
		return
	}

	// Crop to the Backfill Tolerance Value if needed.
	if se.ExtentList[len(se.ExtentList)-1].End.After(t) {
		se.CropToRange(timeseries.Extent{Start: se.ExtentList[0].Start, End: t})
	}

	ts := map[int64]struct{}{}
	for _, dp := range se.Data {
		ts[dp.Time.Unix()] = struct{}{}
	}

	if len(se.Data) == 0 || len(ts) <= sz {
		return
	}

	rc := len(ts) - sz // removal count
	tsl := []int{}
	for k := range ts {
		tsl = append(tsl, int(k))
	}

	sort.Ints(tsl)
	tsl = tsl[rc:]
	tsm := map[int64]struct{}{}
	for _, t := range tsl {
		tsm[int64(t)] = struct{}{}
	}

	min, max := time.Now().Unix(), int64(0)
	newData := DataPoints{}
	for _, dp := range se.Data {
		t := dp.Time.Unix()
		if _, ok := tsm[t]; ok {
			newData = append(newData, dp)
			if t < min {
				min = t
			}

			if t > max {
				max = t
			}
		}
	}

	se.Data = newData
	se.ExtentList = timeseries.ExtentList{timeseries.Extent{
		Start: time.Unix(min, 0),
		End:   time.Unix(max, 0),
	}}

	se.Sort()
}

// Sort sorts all data in the Timeseries chronologically by their timestamp.
func (se *SeriesEnvelope) Sort() {
	sort.Sort(se.Data)
}

// MarshalTimeseries converts a Timeseries into a JSON blob for cache storage.
func (c *Client) MarshalTimeseries(ts timeseries.Timeseries) ([]byte, error) {
	return json.Marshal(ts)
}

// UnmarshalTimeseries converts a JSON blob into a Timeseries value.
func (c *Client) UnmarshalTimeseries(data []byte) (timeseries.Timeseries,
	error) {
	if strings.Contains(strings.Replace(string(data), " ", "", -1),
		`"version":"DF4"`) {
		se := &DF4SeriesEnvelope{}
		err := json.Unmarshal(data, &se)
		return se, err
	}

	se := &SeriesEnvelope{}
	err := json.Unmarshal(data, &se)
	return se, err
}

// UnmarshalInstantaneous is not used for IRONdb origins and is here to conform
// to the Client interface.
func (c Client) UnmarshalInstantaneous(
	data []byte) (timeseries.Timeseries, error) {
	return c.UnmarshalTimeseries(data)
}

// Size returns the approximate memory utilization in bytes of the timeseries
func (se *SeriesEnvelope) Size() int {

	// TODO this implementation is a rough approximation to ensure we conform to the
	// interface specification, it requires refinement in order to be in the ballpark
	c := len(se.Data) * 24
	return c
}
