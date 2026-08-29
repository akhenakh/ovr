// The darwin/arm64 cross build (goreleaser-cross osxcross) has no
// compiler-rt builtins archive to link against, so the availability
// check emitted by clang for @available / __builtin_available cannot be
// resolved. Provide it ourselves: it reports the running macOS version.
#include <sys/sysctl.h>
#include <stdio.h>

#define PLATFORM_MACOS 1

int __isPlatformVersionAtLeast(int variant, int major, int minor, int patch);

int __isPlatformVersionAtLeast(int variant, int major, int minor, int patch) {
	if (variant != PLATFORM_MACOS) {
		return 0;
	}

	char version[32];
	size_t size = sizeof(version);
	if (sysctlbyname("kern.osproductversion", version, &size, NULL, 0) != 0) {
		return 0;
	}

	int osMajor = 0, osMinor = 0, osPatch = 0;
	sscanf(version, "%d.%d.%d", &osMajor, &osMinor, &osPatch);
	if (osMajor != major) {
		return osMajor > major;
	}
	if (osMinor != minor) {
		return osMinor > minor;
	}
	return osPatch >= patch;
}
