package inx

import (
	"context"
	"net"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"

	inx "github.com/iotaledger/inx/go"
	iotago "github.com/iotaledger/iota.go/v3"
)

const (
	workerCount     = 1
	workerQueueSize = 10000
)

func newINXServer() *INXServer {
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	s := &INXServer{grpcServer: grpcServer}
	inx.RegisterINXServer(grpcServer, s)
	return s
}

type INXServer struct {
	inx.UnimplementedINXServer
	grpcServer *grpc.Server
}

func (s *INXServer) ConfigurePrometheus() {
	grpc_prometheus.Register(s.grpcServer)
}

func (s *INXServer) Start() {
	go func() {
		lis, err := net.Listen("tcp", deps.NodeConfig.String(CfgINXBindAddress))
		if err != nil {
			Plugin.LogFatalf("failed to listen: %v", err)
		}
		defer lis.Close()

		if err := s.grpcServer.Serve(lis); err != nil {
			Plugin.LogFatalf("failed to serve: %v", err)
		}
	}()
}

func (s *INXServer) Stop() {
	s.grpcServer.Stop()
}

func (s *INXServer) ReadNodeStatus(context.Context, *inx.NoParams) (*inx.NodeStatus, error) {

	pruningIndex := deps.Storage.SnapshotInfo().PruningIndex

	index, err := deps.UTXOManager.ReadLedgerIndexWithoutLocking()
	if err != nil {
		return nil, err
	}

	latestMilestoneIndex := deps.SyncManager.LatestMilestoneIndex()
	var lmi *inx.MilestoneInfo
	if latestMilestoneIndex > pruningIndex {
		milestone, err := milestoneForIndex(latestMilestoneIndex)
		if err != nil {
			return nil, err
		}
		lmi = milestone.GetMilestoneInfo()
	} else {
		lmi = &inx.MilestoneInfo{
			MilestoneIndex: uint32(latestMilestoneIndex),
		}
	}

	confirmedMilestoneIndex := deps.SyncManager.ConfirmedMilestoneIndex()
	var cmi *inx.MilestoneInfo
	if confirmedMilestoneIndex > pruningIndex {
		milestone, err := milestoneForIndex(confirmedMilestoneIndex)
		if err != nil {
			return nil, err
		}
		cmi = milestone.GetMilestoneInfo()
	} else {
		cmi = &inx.MilestoneInfo{
			MilestoneIndex: uint32(confirmedMilestoneIndex),
		}
	}

	return &inx.NodeStatus{
		IsHealthy:          deps.Tangle.IsNodeHealthy(),
		LatestMilestone:    lmi,
		ConfirmedMilestone: cmi,
		PruningIndex:       uint32(pruningIndex),
		LedgerIndex:        uint32(index),
	}, nil
}

func (s *INXServer) ReadProtocolParameters(context.Context, *inx.NoParams) (*inx.ProtocolParameters, error) {

	var keyRanges []*inx.MilestoneKeyRange
	for _, r := range deps.KeyManager.KeyRanges() {
		keyRanges = append(keyRanges, &inx.MilestoneKeyRange{
			PublicKey:  r.PublicKey[:],
			StartIndex: uint32(r.StartIndex),
			EndIndex:   uint32(r.EndIndex),
		})
	}

	return &inx.ProtocolParameters{
		NetworkName:     deps.NetworkIDName,
		ProtocolVersion: iotago.ProtocolVersion,
		Bech32HRP:       string(deps.Bech32HRP),
		MinPoWScore:     float32(deps.MinPoWScore),
		RentStructure: &inx.RentStructure{
			VByteCost:       deps.DeserializationParameters.RentStructure.VByteCost,
			VByteFactorData: uint64(deps.DeserializationParameters.RentStructure.VBFactorData),
			VByteFactorKey:  uint64(deps.DeserializationParameters.RentStructure.VBFactorKey),
		},
		MilestonePublicKeyCount: uint32(deps.MilestonePublicKeyCount),
		MilestoneKeyRanges:      keyRanges,
	}, nil
}
