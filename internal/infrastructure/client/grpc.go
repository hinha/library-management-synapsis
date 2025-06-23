package client

import (
	"context"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewGRPCClient(ctx context.Context, address string) (*grpc.ClientConn, error) {
	conn, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		//grpc.WithBlock(), // wait until ready
	)
	if err != nil {
		log.Error().Err(err).Str("address", address).Msg("failed to dial service")
		return nil, err
	}

	return conn, nil
}
