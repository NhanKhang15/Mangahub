// Package grpcclient bundles the gateway-side gRPC clients for the four
// downstream services (catalog, artist, progress, prefs) so handlers can
// receive a single struct instead of dialling each connection themselves.
package grpcclient

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	artistpb "mangahub-backend/proto/artistpb"
	catalogpb "mangahub-backend/proto/catalogpb"
	prefspb "mangahub-backend/proto/prefspb"
	progresspb "mangahub-backend/proto/progresspb"
)

// Addresses holds the dial targets for each downstream service.
type Addresses struct {
	Catalog  string
	Artist   string
	Progress string
	Prefs    string
}

// Clients exposes one typed client per downstream service plus the underlying
// connections so the gateway can close them cleanly on shutdown.
type Clients struct {
	Catalog  catalogpb.MangaCatalogClient
	Artist   artistpb.ArtistClient
	Progress progresspb.ReadingProgressClient
	Prefs    prefspb.UserPreferencesClient

	conns []*grpc.ClientConn
}

// Dial opens insecure connections to all four services. Connections are lazy:
// grpc.NewClient does not block, so a downstream service that is still
// starting up does not prevent the gateway from booting.
func Dial(addrs Addresses) (*Clients, error) {
	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{
			"methodConfig": [{
				"name": [{}],
				"retryPolicy": {
					"MaxAttempts": 3,
					"InitialBackoff": "0.1s",
					"MaxBackoff": "1s",
					"BackoffMultiplier": 2.0,
					"RetryableStatusCodes": ["UNAVAILABLE"]
				}
			}]
		}`),
	}

	c := &Clients{}

	catalogConn, err := grpc.NewClient(addrs.Catalog, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("dial catalog %s: %w", addrs.Catalog, err)
	}
	c.conns = append(c.conns, catalogConn)
	c.Catalog = catalogpb.NewMangaCatalogClient(catalogConn)

	artistConn, err := grpc.NewClient(addrs.Artist, dialOpts...)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("dial artist %s: %w", addrs.Artist, err)
	}
	c.conns = append(c.conns, artistConn)
	c.Artist = artistpb.NewArtistClient(artistConn)

	progressConn, err := grpc.NewClient(addrs.Progress, dialOpts...)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("dial progress %s: %w", addrs.Progress, err)
	}
	c.conns = append(c.conns, progressConn)
	c.Progress = progresspb.NewReadingProgressClient(progressConn)

	prefsConn, err := grpc.NewClient(addrs.Prefs, dialOpts...)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("dial prefs %s: %w", addrs.Prefs, err)
	}
	c.conns = append(c.conns, prefsConn)
	c.Prefs = prefspb.NewUserPreferencesClient(prefsConn)

	return c, nil
}

func (c *Clients) Close() {
	for _, conn := range c.conns {
		_ = conn.Close()
	}
	c.conns = nil
}
