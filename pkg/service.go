package gidservice

import (
	gidv1 "github.com/servekit/gid-service/gen/gid/v1"
)

// Service is how a consumer holds gid-service regardless of backend: the
// in-process *Handler (module mode) and the gRPC *Client both satisfy it. It
// embeds the generated server interface so the method set tracks the proto
// automatically — no hand-maintained method list here.
type Service interface {
	gidv1.GidServiceServer
}
