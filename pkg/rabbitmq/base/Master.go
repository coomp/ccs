package base

import (
	"github.com/coomp/ccs/pkg/rabbitmq/comm"
)

// MasterService TODO
type MasterService struct {
	clt *RpcClient
}

// Hearbeat TODO
func (svc *MasterService) Hearbeat(req *comm.HeartbeatRequest) (*comm.HeartbeatResponse, error) {
	rsp := &comm.HeartbeatResponse{}
	err := svc.clt.Send("heartbeat", req, rsp)
	return rsp, err
}

// NewMasterService TODO
func NewMasterService(cc *RpcClient) *MasterService {
	return &MasterService{cc}
}
