package planning

import (
	"fmt"
	"sort"
	"strings"

	version448 "promptfulcustomffmpegbuilder/versions/4.4.8"
	version519 "promptfulcustomffmpegbuilder/versions/5.1.9"
	version616 "promptfulcustomffmpegbuilder/versions/6.1.6"
	version703 "promptfulcustomffmpegbuilder/versions/7.0.3"
	version715 "promptfulcustomffmpegbuilder/versions/7.1.5"
	version803 "promptfulcustomffmpegbuilder/versions/8.0.3"
	version812 "promptfulcustomffmpegbuilder/versions/8.1.2"
	"promptfulcustomffmpegbuilder/versions/shared"
)

// LVersionLibraryWorkImplementation connects a version-specific work id to the
// executable Go manipulation hook under /versions/x.x.x/.
type LVersionLibraryWorkImplementation struct {
	Work        LVersionLibraryWork
	Manipulator shared.LibraryPreparationManipulator
	Plan        shared.LibraryPreparationPlan
}

type LVersionLibraryWorkRegistry struct {
	WorkById map[string]LVersionLibraryWorkImplementation
}

func LVersionLibraryWorkRegistryLoad() (LVersionLibraryWorkRegistry, error) {
	registry := LVersionLibraryWorkRegistry{WorkById: map[string]LVersionLibraryWorkImplementation{}}
	versionLists := []map[string]shared.LibraryPreparationManipulator{
		version448.LLibraryPreparationList,
		version519.LLibraryPreparationList,
		version616.LLibraryPreparationList,
		version703.LLibraryPreparationList,
		version715.LLibraryPreparationList,
		version803.LLibraryPreparationList,
		version812.LLibraryPreparationList,
	}
	for _, versionList := range versionLists {
		libraryIds := make([]string, 0, len(versionList))
		for libraryId := range versionList {
			libraryIds = append(libraryIds, libraryId)
		}
		sort.Strings(libraryIds)
		for _, libraryId := range libraryIds {
			if err := registry.LWorkRegister(libraryId, versionList[libraryId]); err != nil {
				return LVersionLibraryWorkRegistry{}, err
			}
		}
	}
	return registry, nil
}

func (registry LVersionLibraryWorkRegistry) LWorkRegister(libraryId string, manipulator shared.LibraryPreparationManipulator) error {
	if manipulator == nil {
		return fmt.Errorf("version-library work for %q has nil manipulator", libraryId)
	}
	plan := shared.NewLibraryPreparationPlan("", libraryId, "")
	manipulator(plan)
	if strings.TrimSpace(plan.FfmpegVersion) == "" {
		return fmt.Errorf("version-library work for %q did not set FFmpeg version", libraryId)
	}
	if strings.TrimSpace(plan.LibraryId) == "" {
		return fmt.Errorf("version-library work for FFmpeg %q has empty library id", plan.FfmpegVersion)
	}
	work := LVersionLibraryWorkFromPreparationPlan(*plan)
	if _, exists := registry.WorkById[work.WorkId]; exists {
		return fmt.Errorf("duplicate version-library work id %q", work.WorkId)
	}
	registry.WorkById[work.WorkId] = LVersionLibraryWorkImplementation{Work: work, Manipulator: manipulator, Plan: *plan}
	return nil
}

func (registry LVersionLibraryWorkRegistry) LWorkResolve(workId string) (LVersionLibraryWorkImplementation, bool) {
	implementation, exists := registry.WorkById[strings.TrimSpace(workId)]
	return implementation, exists
}

func (registry LVersionLibraryWorkRegistry) LWorkResolveByVersionAndLibrary(ffmpegVersion string, libraryId string) (LVersionLibraryWorkImplementation, bool) {
	return registry.LWorkResolve(LCatalogVersionLibraryWorkIdCreate(ffmpegVersion, libraryId))
}

func (registry LVersionLibraryWorkRegistry) LWorksResolve(workIds []string) ([]LVersionLibraryWork, []string) {
	works := []LVersionLibraryWork{}
	missing := []string{}
	for _, workId := range LCatalogStringsUniqueSortedStable(workIds) {
		implementation, exists := registry.LWorkResolve(workId)
		if !exists {
			missing = append(missing, workId)
			continue
		}
		works = append(works, implementation.Work)
	}
	return works, missing
}

func LVersionLibraryWorkFromPreparationPlan(plan shared.LibraryPreparationPlan) LVersionLibraryWork {
	workId := LCatalogVersionLibraryWorkIdCreate(plan.FfmpegVersion, plan.LibraryId)
	return LVersionLibraryWork{
		WorkId:        workId,
		FfmpegVersion: plan.FfmpegVersion,
		LibraryId:     plan.LibraryId,
		GoFilePath:    plan.VersionSpecificGoFile,
		PhaseNames:    LVersionLibraryWorkPhasesFromPreparationPlan(plan),
		Summary:       LVersionLibraryWorkSummaryCreate(plan),
	}
}

func LVersionLibraryWorkPhasesFromPreparationPlan(plan shared.LibraryPreparationPlan) []LLibraryWorkPhaseName {
	if plan.Method == "" {
		return nil
	}
	return []LLibraryWorkPhaseName{
		LLibraryWorkPhaseAfterLibrarySourceExtract,
		LLibraryWorkPhaseBeforeLibraryConfigure,
		LLibraryWorkPhaseLibraryConfigure,
		LLibraryWorkPhaseLibraryBuild,
		LLibraryWorkPhaseLibraryInstall,
		LLibraryWorkPhaseAfterLibraryInstall,
		LLibraryWorkPhaseBeforeFFmpegConfigure,
	}
}

func LVersionLibraryWorkSummaryCreate(plan shared.LibraryPreparationPlan) string {
	parts := []string{}
	if plan.BuildSystem != "" {
		parts = append(parts, plan.BuildSystem)
	}
	if plan.Method != "" {
		parts = append(parts, plan.Method)
	}
	if plan.PkgConfigName != "" {
		parts = append(parts, "pkg-config "+plan.PkgConfigName)
	}
	if len(parts) == 0 {
		return "version/library manipulation"
	}
	return strings.Join(parts, "; ")
}
