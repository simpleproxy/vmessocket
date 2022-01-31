//go:build !confonly
// +build !confonly

package core

import (
	"io"
	"strings"

	"google.golang.org/protobuf/proto"

	"github.com/vmessocket/vmessocket/common"
	"github.com/vmessocket/vmessocket/common/buf"
	"github.com/vmessocket/vmessocket/common/cmdarg"
	"github.com/vmessocket/vmessocket/main/confloader"
)

type ConfigFormat struct {
	Name      string
	Extension []string
	Loader    ConfigLoader
}

type ConfigLoader func(input interface{}) (*Config, error)

var (
	configLoaderByName = make(map[string]*ConfigFormat)
	configLoaderByExt  = make(map[string]*ConfigFormat)
)

func RegisterConfigLoader(format *ConfigFormat) error {
	name := strings.ToLower(format.Name)
	if _, found := configLoaderByName[name]; found {
		return newError(format.Name, " already registered.")
	}
	configLoaderByName[name] = format

	for _, ext := range format.Extension {
		lext := strings.ToLower(ext)
		if f, found := configLoaderByExt[lext]; found {
			return newError(ext, " already registered to ", f.Name)
		}
		configLoaderByExt[lext] = format
	}

	return nil
}

func getExtension(filename string) string {
	idx := strings.LastIndexByte(filename, '.')
	if idx == -1 {
		return ""
	}
	return filename[idx+1:]
}

func LoadConfig(formatName string, filename string, input interface{}) (*Config, error) {
	ext := getExtension(filename)
	if len(ext) > 0 {
		if f, found := configLoaderByExt[ext]; found {
			return f.Loader(input)
		}
	}

	if f, found := configLoaderByName[formatName]; found {
		return f.Loader(input)
	}

	return nil, newError("Unable to load config in ", formatName).AtWarning()
}

func loadProtobufConfig(data []byte) (*Config, error) {
	config := new(Config)
	if err := proto.Unmarshal(data, config); err != nil {
		return nil, err
	}
	return config, nil
}

func init() {
	common.Must(RegisterConfigLoader(&ConfigFormat{
		Name:      "Protobuf",
		Extension: []string{"pb"},
		Loader: func(input interface{}) (*Config, error) {
			switch v := input.(type) {
			case cmdarg.Arg:
				r, err := confloader.LoadConfig(v[0])
				common.Must(err)
				data, err := buf.ReadAllToBytes(r)
				common.Must(err)
				return loadProtobufConfig(data)
			case io.Reader:
				data, err := buf.ReadAllToBytes(v)
				common.Must(err)
				return loadProtobufConfig(data)
			default:
				return nil, newError("unknow type")
			}
		},
	}))
}
