package planning

import "sort"

func selectConfigureOptions(selectedOptionIds []string) ([]ConfigureOptionChoice, []string) {
	catalog := ConfigureOptionCatalog()
	catalogById := map[string]ConfigureOptionChoice{}
	for _, option := range catalog {
		catalogById[option.OptionId] = option
	}
	selectedOptions := []ConfigureOptionChoice{}
	unknownOptionIds := []string{}
	seen := map[string]bool{}
	for _, selectedOptionId := range selectedOptionIds {
		if selectedOptionId == "" || seen[selectedOptionId] {
			continue
		}
		seen[selectedOptionId] = true
		option, found := catalogById[selectedOptionId]
		if !found {
			unknownOptionIds = append(unknownOptionIds, selectedOptionId)
			continue
		}
		selectedOptions = append(selectedOptions, option)
	}
	return selectedOptions, unknownOptionIds
}

func uniqueFlagsFromConfigureOptions(options []ConfigureOptionChoice) []string {
	flags := []string{}
	seen := map[string]bool{}
	for _, option := range options {
		for _, flag := range option.ConfigureFlags {
			if !seen[flag] {
				flags = append(flags, flag)
				seen[flag] = true
			}
		}
	}
	return flags
}

func selectLibraries(windowsShellProfileName string, selectedLibraryIds []string) ([]LibraryChoice, []string) {
	catalog := LibraryCatalogForShellProfile(windowsShellProfileName)
	catalogById := map[string]LibraryChoice{}
	for _, library := range catalog {
		catalogById[library.LibraryId] = library
	}
	selectedLibraries := []LibraryChoice{}
	unknownLibraryIds := []string{}
	seen := map[string]bool{}
	for _, selectedLibraryId := range selectedLibraryIds {
		if selectedLibraryId == "" || seen[selectedLibraryId] {
			continue
		}
		seen[selectedLibraryId] = true
		library, found := catalogById[selectedLibraryId]
		if !found {
			unknownLibraryIds = append(unknownLibraryIds, selectedLibraryId)
			continue
		}
		selectedLibraries = append(selectedLibraries, library)
	}
	return selectedLibraries, unknownLibraryIds
}

// librariesForConfigureFlags returns catalog entries whose configure flags overlap
// with the given list, excluding any already in skip. Used to resolve ExtraConfigureFlags
// back to their MSYS2 packages.
func librariesForConfigureFlags(windowsShellProfileName string, flags []string, skip []LibraryChoice) []LibraryChoice {
	flagSet := map[string]bool{}
	for _, f := range flags {
		flagSet[f] = true
	}
	skipIds := map[string]bool{}
	for _, lib := range skip {
		skipIds[lib.LibraryId] = true
	}
	result := []LibraryChoice{}
	seen := map[string]bool{}
	for _, lib := range LibraryCatalogForShellProfile(windowsShellProfileName) {
		if skipIds[lib.LibraryId] || seen[lib.LibraryId] {
			continue
		}
		for _, f := range lib.ConfigureFlags {
			if flagSet[f] {
				seen[lib.LibraryId] = true
				result = append(result, lib)
				break
			}
		}
	}
	return result
}

func uniquePackagesFromLibraries(libraries []LibraryChoice) []string {
	packages := []string{}
	seen := map[string]bool{}
	for _, library := range libraries {
		for _, packageName := range library.PackageNames {
			if !seen[packageName] {
				packages = append(packages, packageName)
				seen[packageName] = true
			}
		}
	}
	sort.Strings(packages)
	return packages
}

func uniqueFlagsFromLibraries(libraries []LibraryChoice) []string {
	flags := []string{}
	seen := map[string]bool{}
	for _, library := range libraries {
		for _, flag := range library.ConfigureFlags {
			if !seen[flag] {
				flags = append(flags, flag)
				seen[flag] = true
			}
		}
	}
	return flags
}

func mergeUniqueStrings(first []string, second []string) []string {
	merged := []string{}
	seen := map[string]bool{}
	for _, value := range append(first, second...) {
		if value == "" || seen[value] {
			continue
		}
		merged = append(merged, value)
		seen[value] = true
	}
	return merged
}

func addLicenseFlags(configureFlags []string, licenseProfileName string, libraries []LibraryChoice) []string {
	needsGpl := false
	needsNonfree := licenseProfileName == "nonfree-local"
	needsVersion3 := false
	for _, library := range libraries {
		switch library.LicenseEffectName {
		case "gpl":
			needsGpl = true
		case "nonfree":
			needsNonfree = true
		}
		if libraryRequiresVersion3(library.LibraryId) {
			needsVersion3 = true
		}
	}
	if licenseProfileName == "gpl-local" {
		needsGpl = true
	}
	if needsGpl {
		configureFlags = mergeUniqueStrings([]string{"--enable-gpl"}, configureFlags)
	}
	if needsVersion3 {
		configureFlags = mergeUniqueStrings(configureFlags, []string{"--enable-version3"})
	}
	if needsNonfree {
		configureFlags = mergeUniqueStrings(configureFlags, []string{"--enable-nonfree"})
	}
	return configureFlags
}

func libraryRequiresVersion3(libraryId string) bool {
	switch libraryId {
	case "opencore-amr", "vo-amrwbenc", "lensfun", "aribb24":
		return true
	default:
		return false
	}
}

func deriveLicenseProfileName(selectedLibraries []LibraryChoice, configureFlags []string) string {
	needsGpl := false
	needsNonfree := false
	for _, library := range selectedLibraries {
		switch library.LicenseEffectName {
		case "gpl":
			needsGpl = true
		case "nonfree":
			needsNonfree = true
		}
	}
	for _, configureFlag := range configureFlags {
		switch configureFlag {
		case "--enable-nonfree":
			needsNonfree = true
		case "--enable-gpl":
			needsGpl = true
		}
	}
	if needsNonfree {
		return "nonfree-local"
	}
	if needsGpl {
		return "gpl-local"
	}
	return "lgpl-local"
}

func selectedLibrariesRequireVersion3(selectedLibraries []LibraryChoice) bool {
	for _, library := range selectedLibraries {
		if libraryRequiresVersion3(library.LibraryId) {
			return true
		}
	}
	return false
}
