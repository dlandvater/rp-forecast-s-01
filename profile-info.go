package main

// profile defaults to category and location if empty
// missing profile percentages default to flat year
func getProfile(skuP *SKU) (*Profile, error) {

	var ProfileInfo = new(Profile)
	var err error

	//get profile row
	ProfileInfo = queryOneProfile(skuP)

	//Normalize if needed
	var ttl float32
	for i := 0; i < len(ProfileInfo.ShiftedWeeklyPcnt); i++ {
		ttl = ttl + ProfileInfo.ShiftedWeeklyPcnt[i]
	}
	//Check tolerances
	if ttl < 0.99 || ttl > 1.01 {
		ratio := 1 / ttl
		for i := 1; i < len(ProfileInfo.ShiftedWeeklyPcnt); i++ {
			ProfileInfo.ShiftedWeeklyPcnt[i] = ratio * ProfileInfo.ShiftedWeeklyPcnt[i]
		}
	}
	return ProfileInfo, err
}

// Shift profile percentages to the current week
// For example, profile percentages start at week 1 and extend to week 52
// If the current date is week 10, then create a new slice of profile percentages starting at 10
func shift(offset int, weeklyPcnt []float32) []float32 {

	var ShiftedWeeklyPcnt = make([]float32, 52)
	var j int

	for i := 0; i < 52; i++ {
		j = i + offset
		if j > 51 {
			j = j - 52
		}
		ShiftedWeeklyPcnt[i] = weeklyPcnt[j]
	}
	return ShiftedWeeklyPcnt
}
