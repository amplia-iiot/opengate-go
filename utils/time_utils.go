package utils

import (
	"strconv"
	"strings"
	"time"
)

const ONE_DAY_MS int64 = 24 * 3600 * 1000

func ParseTime2Iso(timeGO time.Time) string {
	return timeGO.Format("2006-01-02T15:04:05.000-07:00")
}
func ParseStrTimeToEpochMS(layout, strTime string) (int64, error) {
	t, err := time.Parse(layout, strTime)
	if err != nil {
		return 0, err
	}
	return t.UnixMilli(), nil
}

// func made available inside .js scripts. Parse human time to ISO format
func GetTime(rawtime string) (string, error) {
	rawtime = strings.ReplaceAll(rawtime, ",", "")
	t, err := time.Parse("Jan 2 2006 15:04:05", rawtime)
	if err != nil {
		return "", err
	}

	formatted := t.Format("2006-01-02T15:04:05.000")
	return formatted, nil
}

// func made available inside .js scripts. Return the seconds between two dates
func GetDiffTime(current, latest string) (float64, error) {
	currentDate, err := time.Parse("06-01-02 15:04:05", current)
	if err != nil {
		return 0.0, err
	}
	latestDate, err := time.Parse("06-01-02 15:04:05", latest)
	if err != nil {
		return 0.0, err
	}
	diff := currentDate.Sub(latestDate).Seconds()
	return diff, nil
}

// func made available inside .js scripts. Parse epoch time to ISO
func ParseEpoch2Iso(rawtime int64) string {
	return time.Unix(rawtime, 0).Format("2006-01-02T15:04:05.000-07:00")
}
func ParseEpoch2IsoMs(rawtime int64) string {
	return time.UnixMilli(rawtime).Format("2006-01-02T15:04:05.000-07:00")
}
func ToGolangTime(strTime, layout string) (time.Time, error) {
	return time.Parse(layout, strTime)
}
func ParseEpochWithLayout(epochTime int64, layout string) string {
	return time.Unix(epochTime, 0).Format(layout)
}
func ParseEpochWithLayoutMs(epochTime int64, layout string) string {
	return time.UnixMilli(epochTime).Format(layout)
}

// convierte cualquier numero en formato epoch y lo pasa a ms
func ParseIntEpochNumber2Ms(numeroEpoch uint64) (numInMs uint64) {
	// Convertir el número Epoch a cadena
	numeroCadena := strconv.FormatUint(numeroEpoch, 10)

	// Obtener la longitud de la cadena
	longitud := len(numeroCadena)

	// Determinar la escala de tiempo según la longitud
	if longitud <= 11 {
		numInMs = numeroEpoch * (1e3)
	} else if longitud <= 14 {
		numInMs = numeroEpoch
	} else if longitud <= 17 {
		numInMs = numeroEpoch / (1e3)
	} else if longitud <= 20 {
		numInMs = numeroEpoch / (1e6)
	}
	return
}

// https://nsidc.org/data/icesat/glas-date-conversion-tool/date_convert/
// Obtiene el tiempo en MS desde la ref. Ejem: j2000(hc): ref=time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
func GetRefMs(refTime time.Time) int64 {
	nowTime := time.Now()
	return int64(nowTime.UnixMilli() - refTime.UnixMilli())
}

// desde un TS  de ref, por ejemplo tipo j2000 obtiene el ts en formato estandar en ms
// ej.: input: 760519359000 (ms in j2000) -- output: 1707204159000 (epoch ts en ms)
// ej: refTime: j2000(hc): refTime=time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
func GetStandardTsMsFromRef(refTime time.Time, tsMsInRef int64) int64 {
	tsTime := refTime.Add(time.Duration(int64(tsMsInRef)) * time.Millisecond)
	return tsTime.UnixMilli()
}

// ts in seconds
func GetStartDayFromEpoch(ts int64) int64 {
	layout := "20060102"
	tsStr := time.Unix(ts, 0).Format(layout)
	t, err := time.Parse(layout, tsStr)
	if err != nil {
		return 0
	}
	return t.Unix()
}