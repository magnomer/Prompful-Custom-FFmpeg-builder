package program

import (
	"errors"

	"promptfulcustomffmpegbuilder/internal/planning"
)

func LHashToolchainVerify(plan planning.LPlanToolchain) error {
	planWithoutHash := plan
	originalPlanHash := planWithoutHash.PlanHash
	planWithoutHash.PlanHash = ""
	computedPlanHash, err := planning.LPlanHashCreate(planWithoutHash)
	if err != nil {
		return err
	}
	if computedPlanHash != originalPlanHash {
		return errors.New("toolchain plan hash does not match plan content")
	}
	return nil
}

func LHashFfmpegVerify(plan planning.LPlanFfmpeg) error {
	planWithoutHash := plan
	originalPlanHash := planWithoutHash.PlanHash
	planWithoutHash.PlanHash = ""
	computedPlanHash, err := planning.LPlanHashCreate(planWithoutHash)
	if err != nil {
		return err
	}
	if computedPlanHash != originalPlanHash {
		return errors.New("FFmpeg build plan hash does not match plan content")
	}
	return nil
}
