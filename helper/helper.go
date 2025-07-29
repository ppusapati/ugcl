package helper

import (
	"encoding/json"
	"math"
	"sort"
	"strconv"

	"gorm.io/datatypes"
	"p9e.in/ugcl/models/kpis"
)

func ToFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func SafeDiv(a, b float64) float64 {
	if math.Abs(b) < 1e-8 {
		return 0.0
	}
	return a / b
}

func MapToKeyValue(m map[string]float64) []kpis.KeyValue {
	list := make([]kpis.KeyValue, 0, len(m))
	for k, v := range m {
		list = append(list, kpis.KeyValue{Key: k, Value: v})
	}
	return list
}

func ToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
func Round(f float64, p int) float64 {
	pow := math.Pow(10, float64(p))
	return math.Round(f*pow) / pow
}
func IfZeroFloat(val float64, d int) float64 {
	if d == 0 {
		return 0.0
	}
	return val
}
func AsStringArray(j datatypes.JSON) []string {
	var arr []string
	_ = json.Unmarshal(j, &arr)
	return arr
}

// Helpers
func IfZero(val float64, d float64) float64 {
	if d == 0 {
		return 0
	} else {
		return val
	}
}
func Percent(num, denom int) float64 {
	if denom == 0 {
		return 0
	} else {
		return float64(num) / float64(denom) * 100
	}
}
func KvpCount(m map[string]int) []kpis.KVP {
	res := make([]kpis.KVP, 0, len(m))
	for k, v := range m {
		res = append(res, kpis.KVP{Key: k, Value: float64(v)})
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Value > res[j].Value })
	return res
}
