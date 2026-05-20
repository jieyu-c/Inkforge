package prompt

import (
	"context"
	"errors"
	"time"

	"github.com/jieyuc/inkforge/services/console/internal/model"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetChannelPointerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetChannelPointerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetChannelPointerLogic {
	return &GetChannelPointerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetChannelPointerLogic) GetChannelPointer() (*types.ChannelPointerResp, error) {
	ns, err := resolveTenantNs(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	key, err := pathPromptKey(l.ctx)
	if err != nil {
		return nil, err
	}
	ch, err := pathChannel(l.ctx)
	if err != nil {
		return nil, err
	}
	pRow, err := loadPromptScoped(l.ctx, l.svcCtx, ns, key)
	if err != nil {
		return nil, err
	}
	ptr, err := l.svcCtx.ConsolePromptChannelPointersModel.FindOneByPromptIdChannelSlug(l.ctx, pRow.Id, ch)
	if errors.Is(err, model.ErrNotFound) {
		return nil, apperr.NotFound("POINTER_NOT_SET", "No published version for this channel")
	}
	if err != nil {
		return nil, err
	}
	if ptr.TenantId != ns.TenantId || ptr.NsId != ns.Id {
		return nil, apperr.Forbidden("POINTER_SCOPE", "Channel pointer is out of scope")
	}
	ver, err := l.svcCtx.ConsolePromptVersionsModel.FindOne(l.ctx, ptr.VersionId)
	if errors.Is(err, model.ErrNotFound) {
		return nil, apperr.NotFound("VERSION_MISSING", "Referenced version no longer exists")
	}
	if err != nil {
		return nil, err
	}
	if ver.PromptId != pRow.Id {
		return nil, apperr.Forbidden("VERSION_SCOPE", "Version does not belong to this prompt")
	}
	return &types.ChannelPointerResp{
		Channel:   ch,
		VersionId: ptr.VersionId,
		Version:   ver.VersionLabel,
		UpdatedAt: ptr.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}
