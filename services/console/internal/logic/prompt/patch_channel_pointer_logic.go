package prompt

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jieyuc/inkforge/services/console/internal/model"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/apperr"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsaudit"
	"github.com/jieyuc/inkforge/services/console/internal/pkg/nsguard"
	"github.com/jieyuc/inkforge/services/console/internal/svc"
	"github.com/jieyuc/inkforge/services/console/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type PatchChannelPointerLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPatchChannelPointerLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PatchChannelPointerLogic {
	return &PatchChannelPointerLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PatchChannelPointerLogic) PatchChannelPointer(req *types.PatchChannelPointerReq) (*types.ChannelPointerResp, error) {
	ns, err := resolveTenantNs(l.ctx, l.svcCtx)
	if err != nil {
		return nil, err
	}
	if err := nsguard.RejectArchivedWrite(ns, "changing channel pointers is not allowed"); err != nil {
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
	vid := strings.TrimSpace(req.VersionId)
	if vid == "" {
		return nil, apperr.BadRequest("INVALID_PAYLOAD", "version_id is required")
	}
	pRow, err := loadPromptScoped(l.ctx, l.svcCtx, ns, key)
	if err != nil {
		return nil, err
	}
	ver, err := l.svcCtx.ConsolePromptVersionsModel.FindOne(l.ctx, vid)
	if errors.Is(err, model.ErrNotFound) {
		return nil, apperr.NotFound("VERSION_NOT_FOUND", "Version does not exist")
	}
	if err != nil {
		return nil, err
	}
	if ver.TenantId != ns.TenantId || ver.NsId != ns.Id || ver.PromptId != pRow.Id {
		return nil, apperr.Forbidden("VERSION_SCOPE", "Version does not belong to this prompt/namespace")
	}
	var fromLabel string
	existing, err := l.svcCtx.ConsolePromptChannelPointersModel.FindOneByPromptIdChannelSlug(l.ctx, pRow.Id, ch)
	if err == nil && existing.TenantId == ns.TenantId {
		fromLabel = existing.VersionId
		newRow := *existing
		newRow.VersionId = vid
		if err := l.svcCtx.ConsolePromptChannelPointersModel.Update(l.ctx, &newRow); err != nil {
			return nil, err
		}
	} else if errors.Is(err, model.ErrNotFound) {
		row := model.ConsolePromptChannelPointers{
			Id:          uuid.NewString(),
			TenantId:    ns.TenantId,
			NsId:        ns.Id,
			PromptId:    pRow.Id,
			ChannelSlug: ch,
			VersionId:   vid,
		}
		if _, err := l.svcCtx.ConsolePromptChannelPointersModel.Insert(l.ctx, &row); err != nil {
			return nil, err
		}
		fromLabel = ""
	} else if err != nil {
		return nil, err
	}
	nsaudit.PromptPointerChanged(ns.TenantId, ns.NsSlug, key, ch, fromLabel, vid)
	ptr, err := l.svcCtx.ConsolePromptChannelPointersModel.FindOneByPromptIdChannelSlug(l.ctx, pRow.Id, ch)
	if err != nil {
		return nil, err
	}
	return &types.ChannelPointerResp{
		Channel:   ch,
		VersionId: ptr.VersionId,
		Version:   ver.VersionLabel,
		UpdatedAt: ptr.UpdatedAt.UTC().Format(time.RFC3339),
	}, nil
}
