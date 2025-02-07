package v2

import (
	"encoding/json"

	"github.com/gohornet/hornet/pkg/model/milestone"
	"github.com/gohornet/hornet/pkg/model/storage"
	"github.com/gohornet/hornet/pkg/model/utxo"
	"github.com/gohornet/hornet/pkg/protocol/gossip"
	iotago "github.com/iotaledger/iota.go/v3"
)

type protocolParameters struct {
	// The Name of the network from which the networkId is derived.
	NetworkName string `json:"networkName"`
	// The protocol version this node supports
	ProtocolVersion byte `json:"protocolVersion"`
	// The Bech32 HRP used.
	Bech32HRP string `json:"bech32HRP"`
	// The minimum pow score of the network.
	MinPoWScore float64 `json:"minPoWScore"`
	// The rent structure according to TIP-19
	RentStructure *iotago.RentStructure `json:"rentStructure"`
}

type nodeStatus struct {
	// Whether the node is healthy.
	IsHealthy bool `json:"isHealthy"`
	// The timestamp of the latest known milestone.
	LatestMilestoneTimestamp uint32 `json:"latestMilestoneTimestamp"`
	// The latest known milestone index.
	LatestMilestoneIndex milestone.Index `json:"latestMilestoneIndex"`
	// The current confirmed milestone's index.
	ConfirmedMilestoneIndex milestone.Index `json:"confirmedMilestoneIndex"`
	// The milestone index at which the last pruning commenced.
	PruningIndex milestone.Index `json:"pruningIndex"`
}

type nodeMetrics struct {
	// The current rate of new messages per second.
	MessagesPerSecond float64 `json:"messagesPerSecond"`
	// The current rate of referenced messages per second.
	ReferencedMessagesPerSecond float64 `json:"referencedMessagesPerSecond"`
	// The ratio of referenced messages in relation to new messages of the last confirmed milestone.
	ReferencedRate float64 `json:"referencedRate"`
}

// infoResponse defines the response of a GET info REST API call.
type infoResponse struct {
	// The name of the node software.
	Name string `json:"name"`
	// The semver version of the node software.
	Version string `json:"version"`
	// The current status of this node.
	Status nodeStatus `json:"status"`
	// The metrics of this node.
	Metrics nodeMetrics `json:"metrics"`
	// The protocol parameters used by this node.
	Protocol protocolParameters `json:"protocol"`
	// The features this node exposes.
	Features []string `json:"features"`
	// The plugins this node exposes.
	Plugins []string `json:"plugins"`
}

// tipsResponse defines the response of a GET tips REST API call.
type tipsResponse struct {
	// The hex encoded message IDs of the tips.
	Tips []string `json:"tipMessageIds"`
}

// receiptsResponse defines the response of a receipts REST API call.
type receiptsResponse struct {
	Receipts []*utxo.ReceiptTuple `json:"receipts"`
}

// messageMetadataResponse defines the response of a GET message metadata REST API call.
type messageMetadataResponse struct {
	// The hex encoded message ID of the message.
	MessageID string `json:"messageId"`
	// The hex encoded message IDs of the parents the message references.
	Parents []string `json:"parentMessageIds"`
	// Whether the message is solid.
	Solid bool `json:"isSolid"`
	// The milestone index that references this message.
	ReferencedByMilestoneIndex *milestone.Index `json:"referencedByMilestoneIndex,omitempty"`
	// If this message represents a milestone this is the milestone index
	MilestoneIndex *milestone.Index `json:"milestoneIndex,omitempty"`
	// The ledger inclusion state of the transaction payload.
	LedgerInclusionState *string `json:"ledgerInclusionState,omitempty"`
	// The reason why this message is marked as conflicting.
	ConflictReason *storage.Conflict `json:"conflictReason,omitempty"`
	// Whether the message should be promoted.
	ShouldPromote *bool `json:"shouldPromote,omitempty"`
	// Whether the message should be reattached.
	ShouldReattach *bool `json:"shouldReattach,omitempty"`
}

// messageCreatedResponse defines the response of a POST messages REST API call.
type messageCreatedResponse struct {
	// The hex encoded message ID of the message.
	MessageID string `json:"messageId"`
}

// childrenResponse defines the response of a GET children REST API call.
type childrenResponse struct {
	// The hex encoded message ID of the message.
	MessageID string `json:"messageId"`
	// The maximum count of results that are returned by the node.
	MaxResults uint32 `json:"maxResults"`
	// The actual count of results that are returned.
	Count uint32 `json:"count"`
	// The hex encoded message IDs of the children of this message.
	Children []string `json:"childrenMessageIds"`
}

// milestoneResponse defines the response of a GET milestones REST API call.
type milestoneResponse struct {
	// The index of the milestone.
	Index uint32 `json:"index"`
	// The hex encoded ID of the message containing the milestone.
	MessageID string `json:"messageId"`
	// The unix time of the milestone payload.
	Time uint32 `json:"timestamp"`
}

// milestoneUTXOChangesResponse defines the response of a GET milestone UTXO changes REST API call.
type milestoneUTXOChangesResponse struct {
	// The index of the milestone.
	Index uint32 `json:"index"`
	// The output IDs (transaction hash + output index) of the newly created outputs.
	CreatedOutputs []string `json:"createdOutputs"`
	// The output IDs (transaction hash + output index) of the consumed (spent) outputs.
	ConsumedOutputs []string `json:"consumedOutputs"`
}

// OutputResponse defines the response of a GET outputs REST API call.
type OutputResponse struct {
	// The hex encoded message ID of the message.
	MessageID string `json:"messageId"`
	// The hex encoded transaction id from which this output originated.
	TransactionID string `json:"transactionId"`
	// The index of the output.
	OutputIndex uint16 `json:"outputIndex"`
	// Whether this output is spent.
	Spent bool `json:"isSpent"`
	// The milestone index at which this output was spent.
	MilestoneIndexSpent milestone.Index `json:"milestoneIndexSpent,omitempty"`
	// The milestone timestamp this output was spent.
	MilestoneTimestampSpent uint32 `json:"milestoneTimestampSpent,omitempty"`
	// The transaction this output was spent with.
	TransactionIDSpent string `json:"transactionIdSpent,omitempty"`
	// The milestone index at which this output was booked into the ledger.
	MilestoneIndexBooked milestone.Index `json:"milestoneIndexBooked"`
	// The milestone timestamp this output was booked in the ledger.
	MilestoneTimestampBooked uint32 `json:"milestoneTimestampBooked"`
	// The ledger index at which this output was available at.
	LedgerIndex milestone.Index `json:"ledgerIndex"`
	// The output in its serialized form.
	RawOutput *json.RawMessage `json:"output,omitempty"`
}

// treasuryResponse defines the response of a GET treasury REST API call.
type treasuryResponse struct {
	MilestoneID string `json:"milestoneId"`
	Amount      string `json:"amount"`
}

// addPeerRequest defines the request for a POST peer REST API call.
type addPeerRequest struct {
	// The libp2p multi address of the peer.
	MultiAddress string `json:"multiAddress"`
	// The alias of the peer.
	Alias *string `json:"alias,omitempty"`
}

// PeerResponse defines the response of a GET peer REST API call.
type PeerResponse struct {
	// The libp2p identifier of the peer.
	ID string `json:"id"`
	// The libp2p multi addresses of the peer.
	MultiAddresses []string `json:"multiAddresses"`
	// The alias of the peer.
	Alias *string `json:"alias,omitempty"`
	// The relation (static, autopeered) of the peer.
	Relation string `json:"relation"`
	// Whether the peer is connected.
	Connected bool `json:"connected"`
	// The gossip protocol information of the peer.
	Gossip *gossip.Info `json:"gossip,omitempty"`
}

// pruneDatabaseRequest defines the request of a prune database REST API call.
type pruneDatabaseRequest struct {
	// The pruning target index.
	Index *milestone.Index `json:"index,omitempty"`
	// The pruning depth.
	Depth *milestone.Index `json:"depth,omitempty"`
	// The target size of the database.
	TargetDatabaseSize *string `json:"targetDatabaseSize,omitempty"`
}

// pruneDatabaseResponse defines the response of a prune database REST API call.
type pruneDatabaseResponse struct {
	// The index of the snapshot.
	Index milestone.Index `json:"index"`
}

// createSnapshotsRequest defines the request of a create snapshots REST API call.
type createSnapshotsRequest struct {
	// The index of the full snapshot.
	FullIndex *milestone.Index `json:"fullIndex,omitempty"`
	// The index of the delta snapshot.
	DeltaIndex *milestone.Index `json:"deltaIndex,omitempty"`
}

// createSnapshotsResponse defines the response of a create snapshots REST API call.
type createSnapshotsResponse struct {
	// The index of the full snapshot.
	FullIndex milestone.Index `json:"fullIndex,omitempty"`
	// The index of the delta snapshot.
	DeltaIndex milestone.Index `json:"deltaIndex,omitempty"`
	// The file path of the full snapshot file.
	FullFilePath string `json:"fullFilePath,omitempty"`
	// The file path of the delta snapshot file.
	DeltaFilePath string `json:"deltaFilePath,omitempty"`
}

// ComputeWhiteFlagMutationsRequest defines the request for a POST debugComputeWhiteFlagMutations REST API call.
type ComputeWhiteFlagMutationsRequest struct {
	// The index of the milestone.
	Index milestone.Index `json:"index"`
	// The timestamp of the milestone.
	Timestamp uint32 `json:"timestamp"`
	// The hex encoded message IDs of the parents the milestone references.
	Parents []string `json:"parentMessageIds"`
	// The hex encoded milestone ID of the previous milestone.
	LastMilestoneID string `json:"lastMilestoneId"`
}

// ComputeWhiteFlagMutationsResponse defines the response for a POST debugComputeWhiteFlagMutations REST API call.
type ComputeWhiteFlagMutationsResponse struct {
	// The hex encoded confirmed merkle tree root as a result of the white flag computation.
	ConfirmedMerkleRoot string `json:"confirmedMerkleRoot"`
	// The hex encoded applied merkle tree root as a result of the white flag computation.
	AppliedMerkleRoot string `json:"appliedMerkleRoot"`
}
