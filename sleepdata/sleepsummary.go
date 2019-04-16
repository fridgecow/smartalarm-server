package sleepdata

import (
	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
	"time"
)

type SleepSummary struct {
	Data       SleepData
	Statistics SleepStatistics

	DataRegions []SleepDataRegion

	timestamps []time.Time

	Title string
}

type SleepRegionType int

const (
	TypeWaking   SleepRegionType = iota
	TypeSleeping                 = iota
	TypeREM                      = iota
)

type SleepDataRegion struct {
	StartTime  time.Time
	EndTime    time.Time
	RegionType SleepRegionType
}

func (summary SleepSummary) GetChartBands() []chart.Series {
	var series []chart.Series

	maxMotion := summary.Data.GetMaxMotion()

	firstWaking := true
	firstREM := true

	for _, region := range summary.DataRegions {
		if region.RegionType == TypeSleeping {
			continue
		}

		x := make([]time.Time, 2)
		y := make([]float64, 2)

		x[0] = region.StartTime
		x[1] = region.EndTime
		y[0] = maxMotion
		y[1] = maxMotion

		color := chart.ColorLightGray
		if region.RegionType == TypeREM {
			color = drawing.ColorFromHex("FFEEB2")
		}

		timeSeries := chart.TimeSeries{
			Style: chart.Style{
				Show:        true,
				FillColor:   color,
				StrokeColor: color,
			},
			XValues: x,
			YValues: y,
		}

		switch region.RegionType {
		case TypeWaking:
			if firstWaking {
				firstWaking = false
				timeSeries.Name = "Light Sleep / Waking"
			}
		case TypeREM:
			if firstREM {
				firstREM = false
				timeSeries.Name = "REM Sleep"
			}
		}

		series = append(series, timeSeries)
	}

	return series
}

func getStartEndTimes(region []string, tz *time.Location) (time.Time, time.Time, error) {
	startMillis, err := ParseFloat(region[1])
	if err != nil {
		return time.Time{}, time.Time{}, Err("could not parse region start time: " + err.Error())
	}

	endMillis, err := ParseFloat(region[2])
	if err != nil {
		return time.Time{}, time.Time{}, Err("could not parse region end time: " + err.Error())
	}

	return time.Unix(int64(startMillis)/1000, 0).In(tz), time.Unix(int64(endMillis)/1000, 0).In(tz), nil

}

func ParseRegions(csv [][]string, tz *time.Location) (summary SleepSummary, e error) {
	start, end, err := getStartEndTimes(csv[0], tz)
	if err != nil {
		return summary, Err("could not establish tracking duration: " + err.Error())
	}
	summary.Statistics.StartTime = start
	summary.Statistics.EndTime = end

	regions := make([]SleepDataRegion, len(csv)*2-3)

	// Compute first region start/end
	start, end, err = getStartEndTimes(csv[1], tz)
	if err != nil {
		return summary, Err("could not parse region: " + err.Error())
	}
	for i, region := range csv[1:] {
		regionType := TypeWaking
		if region[0] == "rem" {
			regionType = TypeREM
			summary.Data.useHRM = true
		}

		regions[2*i] = SleepDataRegion{
			RegionType: regionType,
			StartTime:  start,
			EndTime:    end,
		}

		if 2*i+1 != len(regions) { // If not last
			// Compute next region start, end
			prevEnd := end
			start, end, err = getStartEndTimes(csv[i+2], tz)
			if err != nil {
				return summary, Err("could not parse region: " + err.Error())
			}

			regions[2*i+1] = SleepDataRegion{
				RegionType: TypeSleeping,
				StartTime:  prevEnd,
				EndTime:    start,
			}
		}

	}

	summary.DataRegions = regions
	summary.Statistics.SummaryExport = true
	summary.computeStatistics()

	summary.Title = summary.Statistics.StartTime.Format("Exported Sleep Summary from Monday, January 2")

	return summary, nil
}

func SummariseData(data SleepData) (summary SleepSummary, e error) {
	summary.Data = data

	// Loop through data and construct regions
	var regions []SleepDataRegion
	i := 0
	for i < len(data.Data) {
		dp := data.Data[i]
		j := i

		// Find next point opposite to this one and construct a Sleeping region
		for j < len(data.Data) && data.Data[j].Sleeping == dp.Sleeping {
			j++
		}
		j--

		// Determine region type
		var regionType SleepRegionType
		switch true {
		case dp.Sleeping:
			regionType = TypeSleeping
		case data.GetREMInRegion(i, j):
			regionType = TypeREM
		default:
			regionType = TypeWaking
		}

		// Region is in [i, j]
		regions = append(regions, SleepDataRegion{
			StartTime:  dp.Time,
			EndTime:    data.Data[j].Time,
			RegionType: regionType,
		})

		// Jump to start of next region
		i = j + 1
	}

	// Last region cannot be a REM region
	if regions[len(regions)-1].RegionType == TypeREM {
		regions[len(regions)-1].RegionType = TypeWaking
	}

	summary.Statistics.StartTime = data.Data[0].Time
	summary.Statistics.EndTime = data.Data[len(data.Data)-1].Time

	summary.DataRegions = regions
	summary.computeStatistics()

	summary.Statistics.SummaryExport = false
	summary.Title = summary.Statistics.StartTime.Format("Sleep Summary for Monday, January 2")

	return summary, nil
}

func (summary *SleepSummary) computeStatistics() {
	statistics := &summary.Statistics
	data := &summary.Data
	regions := summary.DataRegions

	statistics.HeartRateEnabled = data.useHRM

	// Find first sleep
	statistics.FirstSleep = statistics.StartTime
	for _, region := range regions {
		if region.RegionType == TypeSleeping || region.RegionType == TypeREM {
			statistics.FirstSleep = region.StartTime
			break
		}
	}

	// Find last wake
	statistics.LastWake = statistics.EndTime
	for i := len(regions) - 1; i >= 0; i-- {
		region := regions[i]
		if region.RegionType == TypeWaking {
			statistics.LastWake = region.StartTime
			break
		}
	}

	if statistics.FirstSleep.After(statistics.LastWake) {
		// Only one sleeping region: just take start and end of tracking
		statistics.FirstSleep = statistics.StartTime
		statistics.LastWake = statistics.EndTime
	}

	// Compute some summary statistics
	var totalDeepTime time.Duration
	var fslwDeepTime time.Duration
	var remTime time.Duration

	for _, region := range regions {
		regionDuration := region.EndTime.Sub(region.StartTime)

		if region.RegionType == TypeSleeping {
			totalDeepTime += regionDuration
			if region.EndTime.After(statistics.FirstSleep) && region.EndTime.Before(statistics.LastWake) {
				fslwDeepTime += regionDuration
			}
		}

		if region.RegionType == TypeREM {
			if region.EndTime.After(statistics.FirstSleep) && region.EndTime.Before(statistics.LastWake) {
				remTime += regionDuration
			}
		}
	}

	totalTime := statistics.EndTime.Sub(statistics.StartTime)
	fslwTime := statistics.LastWake.Sub(statistics.FirstSleep)

	// Sleep Efficiency is portion of time spent in deep sleep
	statistics.SleepEfficiency = float64(totalDeepTime) / float64(totalTime)
	statistics.SleepEfficiencyFSLW = float64(fslwDeepTime) / float64(fslwTime)
	statistics.SleepDuration = fslwTime

	statistics.REMPercent = float64(remTime) / float64(fslwTime)
	statistics.LightPercent = 1 - statistics.REMPercent - statistics.SleepEfficiencyFSLW
}
