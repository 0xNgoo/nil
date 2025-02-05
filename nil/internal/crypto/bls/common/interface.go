package common

// SecretKey represents a BLS secret or private key.
type SecretKey interface {
	PublicKey() PublicKey
	Sign(msg []byte) Signature
	Marshal() []byte
}

// PublicKey represents a BLS public key.
type PublicKey interface {
	// Copy() PublicKey
	// Aggregate(p2 PublicKey) PublicKey
	// IsInfinite() bool
	// Equals(p2 PublicKey) bool
	// Marshal() []byte
}

// Signature represents a BLS signature.
type Signature interface {
	Verify(pubKey PublicKey, msg []byte) bool
	// Deprecated: Use FastAggregateVerify or use this method in spectests only.
	AggregateVerify(pubKeys []PublicKey, msgs [][32]byte) bool
	FastAggregateVerify(pubKeys []PublicKey, msg [32]byte) bool
	Eth2FastAggregateVerify(pubKeys []PublicKey, msg [32]byte) bool
	Marshal() []byte
	Copy() Signature
}
