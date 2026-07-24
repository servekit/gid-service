package gidservice

import (
	"github.com/servekit/go-common/lifecycle"
	"github.com/servekit/go-common/signalx"
)

var (
	_ lifecycle.Service = (*Server)(nil)
	_ signalx.Service   = (*Server)(nil)
)
