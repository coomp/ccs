package app

import (
	"github.com/coomp/ccs/lib/net/pool"
)

type CCSServer struct {
	Pool *pool.Pool
}

func NewCCSServer() *CCSServer {

	return &CCSServer{}
}

func (s *CCSServer) Start() {
}
