package features

import "github.com/vmessocket/vmessocket/common"

//go:generate go run github.com/vmessocket/vmessocket/common/errors/errorgen

type Feature interface {
	common.HasType
	common.Runnable
}

func PrintDeprecatedFeatureWarning(feature string) {
	newError("You are using a deprecated feature: " + feature + ". Please update your config file with latest configuration format, or update your client software.").WriteToLog()
}
