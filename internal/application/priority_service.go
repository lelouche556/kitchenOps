package application

import "time"

func (s *KitchenAppService) queueScore(basePriority float64, createdAt time.Time) float64 {
	waitSecs := time.Since(createdAt).Seconds()
	effective := basePriority + waitSecs*s.agingFactor
	return -effective
}
