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
	version901 "promptfulcustomffmpegbuilder/versions/9.0.1"
	"promptfulcustomffmpegbuilder/versions/shared"
)

// LWorkImplementation connects a version-specific work id to the
// executable Go manipulation hook under /versions/x.x.x/.
type LWorkImplementation struct {
	Work        LVersionLibraryWork
	Manipulator shared.LPreparationManipulator
	Plan        shared.LPreparationPlan
}

type LWorkRegistry struct {
	WorkById map[string]LWorkImplementation
}

func LWorkRegistryLoad() (LWorkRegistry, error) {
	registry := LWorkRegistry{WorkById: map[string]LWorkImplementation{}}
	versionLists := []map[string]shared.LPreparationManipulator{
		version448.LPreparationCatalog,
		version519.LPreparationCatalog,
		version616.LPreparationCatalog,
		version703.LPreparationCatalog,
		version715.LPreparationCatalog,
		version803.LPreparationCatalog,
		version812.LPreparationCatalog,
		version901.LPreparationCatalog,
	}
	for _, versionList := range versionLists {
		libraryIds := make([]string, 0, len(versionList))
		for libraryId := range versionList {
			libraryIds = append(libraryIds, libraryId)
		}
		sort.Strings(libraryIds)
		for _, libraryId := range libraryIds {
			if err := registry.LWorkRegister(libraryId, versionList[libraryId]); err != nil {
				return LWorkRegistry{}, err
			}
		}
	}
	return registry, nil
}

func (registry LWorkRegistry) LWorkRegister(libraryId string, manipulator shared.LPreparationManipulator) error {
	if manipulator == nil {
		return fmt.Errorf("version-library work for %q has nil manipulator", libraryId)
	}
	plan := shared.LPreparationPlanCreate("", libraryId, "")
	manipulator(plan)
	if strings.TrimSpace(plan.FfmpegVersion) == "" {
		return fmt.Errorf("version-library work for %q did not set FFmpeg version", libraryId)
	}
	if strings.TrimSpace(plan.LibraryId) == "" {
		return fmt.Errorf("version-library work for FFmpeg %q has empty library id", plan.FfmpegVersion)
	}
	work := LWorkPlanResolve(*plan)
	if _, exists := registry.WorkById[work.WorkId]; exists {
		return fmt.Errorf("duplicate version-library work id %q", work.WorkId)
	}
	registry.WorkById[work.WorkId] = LWorkImplementation{Work: work, Manipulator: manipulator, Plan: *plan}
	return nil
}

func (registry LWorkRegistry) LWorkResolve(workId string) (LWorkImplementation, bool) {
	implementation, exists := registry.WorkById[strings.TrimSpace(workId)]
	return implementation, exists
}

func (registry LWorkRegistry) LWorkLibraryResolve(ffmpegVersion string, libraryId string) (LWorkImplementation, bool) {
	return registry.LWorkResolve(LWorkIdentifierCreate(ffmpegVersion, libraryId))
}

func (registry LWorkRegistry) LWorksResolve(workIds []string) ([]LVersionLibraryWork, []string) {
	works := []LVersionLibraryWork{}
	missing := []string{}
	for _, workId := range LStringsSortedGet(workIds) {
		implementation, exists := registry.LWorkResolve(workId)
		if !exists {
			missing = append(missing, workId)
			continue
		}
		works = append(works, implementation.Work)
	}
	return works, missing
}

func LWorkPlanResolve(plan shared.LPreparationPlan) LVersionLibraryWork {
	workId := LWorkIdentifierCreate(plan.FfmpegVersion, plan.LibraryId)
	return LVersionLibraryWork{
		WorkId:        workId,
		FfmpegVersion: plan.FfmpegVersion,
		LibraryId:     plan.LibraryId,
		GoFilePath:    plan.VersionSpecificGoFile,
		PhaseNames:    LWorkPhaseResolve(plan),
		Summary:       LWorkSummaryCreate(plan),
	}
}

func LWorkPhaseResolve(plan shared.LPreparationPlan) []LWorkPhaseName {
	if plan.Method == "" {
		return nil
	}
	return []LWorkPhaseName{
		LSourceExtractAfter,
		LLibraryConfigureBefore,
		LLibraryConfigurePhase,
		LLibraryBuildPhase,
		LLibraryInstallPhase,
		LLibraryInstallAfter,
		LFFmpegConfigureBefore,
	}
}

func LWorkSummaryCreate(plan shared.LPreparationPlan) string {
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
