package version812

import (
	"reflect"
	"strings"
	"testing"

	"promptfulcustomffmpegbuilder/internal/scripting"
	"promptfulcustomffmpegbuilder/versions/shared"
)

func TestLibraryVvencPrepareAppendsCxxRuntimeToPkgConfigLibs(t *testing.T) {
	plan := shared.NewLibraryPreparationPlan("8.1.2", "vvenc", "versions/8.1.2/vvenc.go")

	LLibraryVvencPrepare(plan)

	if plan.PkgConfigName != "libvvenc" {
		t.Fatalf("PkgConfigName = %q, want libvvenc", plan.PkgConfigName)
	}
	if !reflect.DeepEqual(plan.PkgConfigAppendLibs, []string{"stdc++"}) {
		t.Fatalf("PkgConfigAppendLibs = %#v, want []string{\"stdc++\"}", plan.PkgConfigAppendLibs)
	}
}

func TestLibraryVvencPrepareScriptPatchesPkgConfigLibsForSharedFfmpeg(t *testing.T) {
	plan := shared.NewLibraryPreparationPlan("8.1.2", "vvenc", "versions/8.1.2/vvenc.go")
	LLibraryVvencPrepare(plan)

	lines, err := scripting.LScriptLibraryInternalCreate(scripting.LLibraryBuildSpec{
		LibraryId:                plan.LibraryId,
		DisplayName:              plan.DisplayName,
		BuildSystem:              plan.BuildSystem,
		CMakeOptions:             plan.CMakeOptions,
		PkgConfigName:            plan.PkgConfigName,
		PkgConfigAppendLibs:      plan.PkgConfigAppendLibs,
		VerifyHeaderRelativePath: plan.VerifyHeaderRelativePath,
		VerifyLibStem:            plan.VerifyLibStem,
	})
	if err != nil {
		t.Fatalf("LScriptLibraryInternalCreate returned error: %v", err)
	}

	script := strings.Join(lines, "\n")
	for _, want := range []string{
		"Patching libvvenc.pc",
		`s/$/ -lstdc++/`,
		`grep '^Libs:'`,
	} {
		if !strings.Contains(script, want) {
			t.Fatalf("generated vvenc prep script does not contain %q:\n%s", want, script)
		}
	}
}
