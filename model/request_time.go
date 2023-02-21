package model

type RequestTimeList []uint64

func (rt RequestTimeList) Len() int {
	return len(rt)
}

func (rt RequestTimeList) Less(i int, j int) bool {
	return rt[i] < rt[j]
}
func (rt RequestTimeList) Swap(i int, j int) {
	rt[i], rt[j] = rt[j], rt[i]
}
