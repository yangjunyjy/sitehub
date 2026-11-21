package health

import (
	"context"

	"sitehub/service/rbac/api/internal/logic"
	"sitehub/service/rbac/api/internal/svc"
	"sitehub/service/rbac/api/internal/types"
	pb "sitehub/service/rbac/rpc/pb/rbac"
)

func Health(ctx context.Context, services *svc.ServiceContext) (*types.Response, error) {
	return logic.Forward(ctx, services, nil, &pb.HealthRequest{}, services.RPC.Health)
}
