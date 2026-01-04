package main

func shiftMatches(shift *Shift) bool {
	durationHours := shift.EndTime.Sub(shift.StartTime).Hours()

	validRole :=
		shift.Role == "LG" ||
			shift.Role == "Deck Coordinator"

	return validRole && durationHours >= 3
}
