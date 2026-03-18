package testdata

// generate:reset
type ResetableStruct struct {
	i     int
	str   string
	strP  *string
	s     []int
	m     map[string]string
	child *ResetableStruct
}

// generate:reset
type ResetableStruct2 struct {
	i     int
	str   string
	child ResetableStruct
}
