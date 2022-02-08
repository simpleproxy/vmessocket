//go:build !windows
// +build !windows

package platform

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

func ExpandEnv(s string) string {
	return os.ExpandEnv(s)
}

func GetAssetLocation(file string) string {
	const name = "vmessocket.location.asset"
	assetPath := NewEnvFlag(name).GetValue(getExecutableDir)
	defPath := filepath.Join(assetPath, file)
	for _, p := range []string{
		defPath,
		filepath.Join("/usr/local/share/vmessocket/", file),
		filepath.Join("/usr/share/vmessocket/", file),
		filepath.Join("/opt/share/vmessocket/", file),
	} {
		if _, err := os.Stat(p); err != nil && errors.Is(err, fs.ErrNotExist) {
			continue
		}
		return p
	}
	return defPath
}

func GetToolLocation(file string) string {
	const name = "vmessocket.location.tool"
	toolPath := EnvFlag{Name: name, AltName: NormalizeEnvName(name)}.GetValue(getExecutableDir)
	return filepath.Join(toolPath, file)
}

func LineSeparator() string {
	return "\n"
}
