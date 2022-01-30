package transport

import "github.com/vmessocket/vmessocket/common/buf"

type Link struct {
	Reader buf.Reader
	Writer buf.Writer
}
