package task

import "github.com/vmessocket/vmessocket/common"

func Close(v interface{}) func() error {
	return func() error {
		return common.Close(v)
	}
}
