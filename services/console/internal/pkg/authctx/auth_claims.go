package authctx

import (
	"context"
	"encoding/json"
	"strconv"
)

// JWT custom claims copied into Go context keys by go-zero Authorize middleware
// (see github.com/zeromicro/go-zero/rest/handler/authhandler.go).

const KeysUserID = "user_id"
const KeysSessionID = "sid"

func StringClaim(ctx context.Context, key string) (string, bool) {
	raw := ctx.Value(key)
	switch v := raw.(type) {
	case string:
		return v, true
	case json.Number:
		return v.String(), true
	case float64:
		return strconv.FormatInt(int64(v), 10), true
	default:
		return "", false
	}
}
