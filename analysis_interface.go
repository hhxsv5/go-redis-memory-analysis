package gorma

type AnalysisInterface interface {
	Start(delimiters []string)
	SaveReports(folder string) error
	Close()
}

type Report struct {
	Key         string
	Count       uint64
	Size        uint64
	NeverExpire uint64
	AvgTtl      uint64
}

type DBReports map[uint64][]Report
type KeyReports map[string]Report
type SortBySizeReports []Report
type SortByCountReports []Report

func (sr SortBySizeReports) Len() int {
	return len(sr)
}

func (sr SortBySizeReports) Less(i, j int) bool {
	return sr[i].Size > sr[j].Size
}

func (sr SortBySizeReports) Swap(i, j int) {
	sr[i], sr[j] = sr[j], sr[i]
}

func (sr SortByCountReports) Len() int {
	return len(sr)
}

func (sr SortByCountReports) Less(i, j int) bool {
	return sr[i].Count > sr[j].Count
}

func (sr SortByCountReports) Swap(i, j int) {
	sr[i], sr[j] = sr[j], sr[i]
}
