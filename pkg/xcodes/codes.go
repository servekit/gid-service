// Package xcodes defines gid-service error codes. Generic codes (ErrInternal,
// ErrBadRequest, etc.) live in go-common's xerr package; service-specific
// codes land here.
//
// Usage:
//
//	if err != nil {
//	    return nil, xcodes.ErrGenerateID.Wrapf(err, "generate id")
//	}
package xcodes

import "github.com/servekit/go-common/xerr"

// gid-service error codes.
var (
	ErrReadConfig      = xerr.New("GID_READ_CONFIG_FAILED", xerr.CategoryInternal, 500, "read config")
	ErrUnmarshalConfig = xerr.New("GID_UNMARSHAL_CONFIG_FAILED", xerr.CategoryInternal, 500, "unmarshal config")
	ErrCreateSonyflake = xerr.New("GID_CREATE_SONYFLAKE_FAILED", xerr.CategoryInternal, 500, "create sonyflake")
	ErrGenerateID      = xerr.New("GID_GENERATE_ID_FAILED", xerr.CategoryInternal, 500, "generate id")
	ErrDialService     = xerr.New("GID_DIAL_SERVICE_FAILED", xerr.CategoryServiceUnavailable, 503, "dial gid service")

	ErrConfigRequired             = xerr.New("GID_CONFIG_REQUIRED", xerr.CategoryBadRequest, 400, "config is required").New()
	ErrServerConfigRequired       = xerr.New("GID_SERVER_CONFIG_REQUIRED", xerr.CategoryBadRequest, 400, "server config is required").New()
	ErrServerGRPCAddrRequired     = xerr.New("GID_SERVER_GRPC_ADDR_REQUIRED", xerr.CategoryBadRequest, 400, "server.grpc_addr is required").New()
	ErrSnowflakeConfigRequired    = xerr.New("GID_SNOWFLAKE_CONFIG_REQUIRED", xerr.CategoryBadRequest, 400, "snowflake config is required").New()
	ErrSnowflakeMachineIDInvalid  = xerr.New("GID_SNOWFLAKE_MACHINE_ID_INVALID", xerr.CategoryBadRequest, 400, "snowflake.machine_id must be between 1 and 65535").New()
	ErrSnowflakeStartTimeRequired = xerr.New("GID_SNOWFLAKE_START_TIME_REQUIRED", xerr.CategoryBadRequest, 400, "snowflake.start_time is required").New()
	ErrSnowflakeStartTimeFuture   = xerr.New("GID_SNOWFLAKE_START_TIME_FUTURE", xerr.CategoryBadRequest, 400, "snowflake.start_time must not be in the future").New()

	ErrMachineIDInvalid  = xerr.New("GID_MACHINE_ID_INVALID", xerr.CategoryBadRequest, 400, "machine_id must be between 1 and 65535").New()
	ErrStartTimeFuture   = xerr.New("GID_START_TIME_FUTURE", xerr.CategoryBadRequest, 400, "start_time must not be in the future").New()
	ErrBatchCountInvalid = xerr.New("GID_BATCH_COUNT_INVALID", xerr.CategoryBadRequest, 400, "count must be between 1 and 1000").New()
)
