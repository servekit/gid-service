package gidservice

import (
	"context"

	gidv1 "github.com/servekit/gid-service/gen/gid/v1"
)

// Service is how a consumer holds gid-service regardless of backend: the
// in-process *Handler (module mode) and the gRPC *Client both satisfy it. It
// embeds the generated server interface so the method set tracks the proto
// automatically — no hand-maintained method list here.
type Service interface {
	gidv1.GidServiceServer
}

// NextID fetches one int64 ID from a gid backend over the proto-shaped
// Service interface, unwrapping the request/response for callers that just
// need the number. It lives with the provider so every consumer shares one
// implementation instead of a per-service helper.
func NextID(ctx context.Context, svc Service) (int64, error) {
	resp, err := svc.NextID(ctx, &gidv1.NextIDRequest{})
	if err != nil {
		return 0, err
	}
	return resp.GetId(), nil
}
