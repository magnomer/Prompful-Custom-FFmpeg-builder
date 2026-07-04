package planning

func LPackageDefaultGet(windowsShellProfileName string) []string {
	packagePrefix := LPackageProfileResolve(windowsShellProfileName)
	return []string{
		"base-devel",
		"git",
		"make",
		"diffutils",
		packagePrefix + "-binutils",
		packagePrefix + "-crt",
		packagePrefix + "-gcc",
		packagePrefix + "-headers",
		packagePrefix + "-libmangle",
		packagePrefix + "-libwinpthread",
		packagePrefix + "-make",
		packagePrefix + "-pkgconf",
		packagePrefix + "-tools",
		packagePrefix + "-winpthreads",
		packagePrefix + "-winstorecompat",
		packagePrefix + "-cmake",
		packagePrefix + "-ninja",
		packagePrefix + "-nasm",
		packagePrefix + "-yasm",
	}
}

func LPackageProfileResolve(windowsShellProfileName string) string {
	switch windowsShellProfileName {
	case "mingw64":
		return "mingw-w64-x86_64"
	case "clang64":
		return "mingw-w64-clang-x86_64"
	case "ucrt64", "":
		return "mingw-w64-ucrt-x86_64"
	default:
		return "mingw-w64-ucrt-x86_64"
	}
}
