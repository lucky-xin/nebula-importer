package manager

import "testing"

func TestSubSlice(t *testing.T) {
	total := 891
	var values []int
	for i := range total {
		values = append(values, i)
	}
	batch := 1000
	length := len(values)
	times := length / batch
	if length%batch != 0 || times == 0 {
		times++
	}
	for j := range times {
		start := j * batch
		end := (j + 1) * batch
		if end > length {
			end = length
		}
		subs := values[start:end]
		l := len(subs)
		println(l, subs[0], subs[l-1])
	}
}
