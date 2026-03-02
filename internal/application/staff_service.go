package application

import (
	"math"
	"time"

	"SwishAssignment/internal/models"
)

func (s *KitchenAppService) pickStaff(staff []models.StaffModel, now time.Time) models.StaffModel {
	best := staff[0]
	bestScore := s.staffScore(best, now)
	for i := 1; i < len(staff); i++ {
		score := s.staffScore(staff[i], now)
		if score < bestScore {
			best = staff[i]
			bestScore = score
		}
	}
	return best
}

func (s *KitchenAppService) staffScore(st models.StaffModel, now time.Time) float64 {
	shiftSecs := int(math.Max(1, now.Sub(st.ShiftStart).Seconds()))
	util := float64(st.ActiveSeconds) / float64(shiftSecs)
	return s.weightLoad*float64(st.ActiveTasks) + s.weightUtil*util - s.weightEff*st.EfficiencyMultiplier
}
