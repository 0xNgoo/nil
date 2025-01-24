package proto

import (
	"github.com/NilFoundation/nil/nil/internal/types"
)

func Uint256ToProtoUint256(u types.Uint256) *Uint256 {
	return &Uint256{
		WordParts: u[:],
	}
}

func ProtoUint256ToUint256(pb *Uint256) types.Uint256 {
	var u types.Uint256
	copy(u[:], pb.WordParts)
	return u
}
