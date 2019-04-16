package sleepdata

import (
	"fmt"
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"math"
	"sort"
	"time"
)

var timeStep = time.Minute
var sleepWeights = [7]int{106, 54, 58, 76, 230, 74, 67}
var sleepProduct = 0.001

type SleepDataPoint struct {
	Time                         time.Time
	Motion                       float64
	HeartRate                    float64
	SDNN                         float64
	HeartRateVariabilityEstimate float64

	SleepDiscriminant float64
	Sleeping          bool
}

type SleepStatistics struct {
	HeartRateEnabled    bool
	SummaryExport       bool
	StartTime           time.Time
	FirstSleep          time.Time
	SleepEfficiency     float64
	SleepEfficiencyFSLW float64
	REMPercent          float64
	LightPercent        float64
	LastWake            time.Time
	EndTime             time.Time

	SleepDuration time.Duration
}

type SleepData struct {
	Data []SleepDataPoint
	tz   time.Location

	useHRM  bool
	useSDNN bool

	remThreshold float64
}

var Err = fmt.Errorf

func (sd SleepData) calculateREMThreshold() float64 {
	if !sd.useHRM {
		return 0
	}

	if sd.useSDNN {
		return 0.19 // A value that seems to work okay
	}

	// Average HR
	avgHR := 0.0
	for _, x := range sd.Data {
		avgHR += x.HeartRate
	}
	avgHR = avgHR / float64(len(sd.Data))

	// Perform a lowpass on HRM to filter out faster than 10x / minute
	lowpass := make([]float64, len(sd.Data))
	state := avgHR
	for i, x := range sd.Data {
		state += 0.15 * (x.HeartRate - state)
		lowpass[i] = state
	}

	// Compute delta from original heart rate
	deltas := make([]float64, len(sd.Data))
	for i, x := range lowpass {
		deltas[i] = math.Abs(x - sd.Data[i].HeartRate)
		sd.Data[i].HeartRateVariabilityEstimate = deltas[i]
	}

	// Filter out any variability below 18
	var filtered []float64
	for _, x := range deltas {
		if x > 18 {
			filtered = append(filtered, x)
		}
	}

	if len(filtered) == 0 {
		return 0
	}

	// Take the 20th percentile
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i] < filtered[j]
	})

	return filtered[int(math.Round(0.8*float64(len(filtered))))]
}

func (sd SleepData) GetMotionSeries() chart.TimeSeries {
	x := make([]time.Time, len(sd.Data))
	y := make([]float64, len(sd.Data))

	for i, dp := range sd.Data {
		x[i] = dp.Time
		y[i] = dp.Motion
	}

	return chart.TimeSeries{
		Name: "Motion",
		Style: chart.Style{
			Show:        true,
			StrokeColor: drawing.ColorFromHex("008080"),
		},
		XValues: x,
		YValues: y,
	}
}

func (sd SleepData) GetHeartRateSeries() chart.TimeSeries {
	if !sd.useHRM {
		return chart.TimeSeries{}
	}

	x := make([]time.Time, len(sd.Data))
	y := make([]float64, len(sd.Data))

	for i, dp := range sd.Data {
		x[i] = dp.Time
		y[i] = dp.HeartRate
	}

	return chart.TimeSeries{
		Name: "Heart Rate",
		Style: chart.Style{
			Show:        true,
			StrokeColor: drawing.ColorFromHex("800000"),
		},
		YAxis:   chart.YAxisSecondary,
		XValues: x,
		YValues: y,
	}
}

func (sd SleepData) GetMaxMotion() float64 {
	if len(sd.Data) == 0 {
		return 1000
	}

	var max float64
	for _, dp := range sd.Data {
		if dp.Motion > max {
			max = dp.Motion
		}
	}

	return max
}

func (sd SleepData) GetREMInRegion(start int, end int) bool {
	if !sd.useHRM {
		return false
	}

	// Compute "any" between i and j
	for i := start; i <= end; i++ {
		if sd.useSDNN {
			if sd.Data[i].SDNN > sd.remThreshold {
				return true
			}
		} else if sd.Data[i].HeartRateVariabilityEstimate > sd.remThreshold {
			return true
		}
	}

	return false
}

func MakeSleepData(data [][]string, tz time.Location) (sleepData SleepData, e error) {
	sleepData.tz = tz

	if len(data) < 2 {
		return sleepData, Err("not enough data for export")
	}

	// Parse [][]string into []SleepDataPoint
	var dataPoints []SleepDataPoint
	for _, point := range data {
		var dp SleepDataPoint

		millis, err := ParseFloat(point[0])
		if err != nil {
			return sleepData, err
		}

		dp.Time = time.Unix(int64(millis)/1000, 0).In(&tz)

		dp.Motion, err = ParseFloat(point[1])
		if err != nil {
			return sleepData, err
		}
		dp.Motion = math.Max(dp.Motion, 0)

		if len(point) < 3 {
			dataPoints = append(dataPoints, dp)
			continue
		}

		dp.HeartRate, err = ParseFloat(point[2])
		if err != nil {
			return sleepData, err
		}
		sleepData.useHRM = true

		if len(point) < 4 {
			dataPoints = append(dataPoints, dp)
			continue
		}

		dp.SDNN, err = ParseFloat(point[3])
		if err != nil {
			return sleepData, err
		}
		sleepData.useSDNN = true

		dataPoints = append(dataPoints, dp)
	}
	sleepData.Data = dataPoints

	// Calculate REM threshold
	sleepData.remThreshold = sleepData.calculateREMThreshold()

	// Calculate new datapoints interpolated at regular intervals
	var interpDataPoints []SleepDataPoint

	startTime := dataPoints[0].Time
	endTime := dataPoints[len(dataPoints)-1].Time
	currentIndex := 0

	for currentTime := startTime; currentTime.Before(endTime); currentTime = currentTime.Add(timeStep) {
		if currentIndex+1 >= len(dataPoints) {
			// Nothing to interpolate
			break
		}

		if currentTime.After(dataPoints[currentIndex+1].Time) {
			// Increment currentIndex to interpolate between next two points
			currentIndex++
		}

		dpLower := dataPoints[currentIndex]
		dpUpper := dataPoints[currentIndex+1]

		secsLower := dpLower.Time.Unix()
		secsUpper := dpUpper.Time.Unix()
		secsTarget := currentTime.Unix()

		ratio := math.Max(
			math.Min(
				float64(secsTarget-secsLower)/float64(secsUpper-secsLower),
				1,
			),
			0,
		)

		interpDataPoints = append(interpDataPoints, SleepDataPoint{
			Time:                         currentTime,
			Motion:                       math.Max(math.Round(dpLower.Motion*(1-ratio)+dpUpper.Motion*ratio), 0),
			HeartRate:                    dpLower.HeartRate*(1-ratio) + dpUpper.HeartRate*ratio,
			SDNN:                         dpLower.SDNN*(1-ratio) + dpUpper.SDNN*ratio,
			HeartRateVariabilityEstimate: dpLower.HeartRateVariabilityEstimate*(1-ratio) + dpUpper.HeartRateVariabilityEstimate*ratio,
		})
	}

	// Calculate sleep discriminants and whether sleeping at each DP
	for i, dp := range interpDataPoints {
		for j := 0; j < 7; j++ {
			if i+j < 4 {
				// too early, assume activity 300 - i.e, not sleeping
				dp.SleepDiscriminant += float64(sleepWeights[j] * 300)
				continue
			}

			if i+j+3 > len(interpDataPoints) {
				// too late, repeat last activity
				activity := math.Min(interpDataPoints[len(interpDataPoints)-1].Motion, 300)
				dp.SleepDiscriminant += activity * float64(sleepWeights[j])
				continue
			}

			dp2 := interpDataPoints[i+j-4]
			activity := math.Min(dp2.Motion, 300)
			dp.SleepDiscriminant += activity * float64(sleepWeights[j])
		}

		dp.SleepDiscriminant *= sleepProduct
		dp.Sleeping = dp.SleepDiscriminant < 1

		interpDataPoints[i] = dp
	}

	sleepData.Data = interpDataPoints
	return sleepData, nil
}
