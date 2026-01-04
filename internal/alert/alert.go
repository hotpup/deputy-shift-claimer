package main

import (
	"fmt"
)

func triggerAlert(shift *Shift) {
	fmt.Println("🚨 MATCHING SHIFT FOUND 🚨")
	fmt.Printf("Role: %s\n", shift.Role)
	fmt.Printf("Time: %s - %s\n",
		shift.StartTime.Format("3:04 PM"),
		shift.EndTime.Format("3:04 PM"),
	)
	fmt.Printf("Duration: %.1f hours\n",
		shift.EndTime.Sub(shift.StartTime).Hours(),
	)
}
