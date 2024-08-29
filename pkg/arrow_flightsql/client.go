package arrow_flightsql

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"sync"

	"github.com/apache/arrow/go/v12/arrow/flight"
	"github.com/apache/arrow/go/v12/arrow/flight/flightsql"
	"github.com/apache/arrow/go/v12/arrow/ipc"
	"github.com/apache/arrow/go/v12/arrow/memory"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// newFlightSQLClient creates a new FlightSQL client using the provided configuration.
func newFlightSQLClient(cfg config) (*client, error) {
	dialOptions, err := grpcDialOptions(cfg)
	if err != nil {
		return nil, errors.Join(errors.New("newFlightSQLClient DialOptions"), err)
	}

	fsqlClient, err := flightsql.NewClient(cfg.Addr, nil, nil, dialOptions...)
	if err != nil {
		return nil, err
	}

	return &client{Client: fsqlClient}, nil
}

// grpcDialOptions returns the gRPC dial options based on the configuration.
func grpcDialOptions(cfg config) ([]grpc.DialOption, error) {
	if cfg.Secure {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("x509: %s", err)
		}
		return []grpc.DialOption{
			grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, "")),
		}, nil
	}

	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}, nil
}

// client wraps a flightsql.Client to extend its behavior and provide access to gRPC headers for streaming operations.
type client struct {
	*flightsql.Client
}

// FlightClient returns the underlying flight.Client.
func (c *client) FlightClient() flight.Client {
	return c.Client.Client
}

// DoGetWithHeaderExtraction performs a DoGet and wraps the stream to extract headers when available.
func (c *client) DoGetWithHeaderExtraction(ctx context.Context, in *flight.Ticket, opts ...grpc.CallOption) (*flightReader, error) {
	stream, err := c.FlightClient().DoGet(ctx, in, opts...)
	if err != nil {
		return nil, err
	}
	return newFlightReader(stream, c.Client.Alloc)
}

// flightReader wraps a flight.Reader to expose the headers captured during the first read.
type flightReader struct {
	*flight.Reader
	extractor *headerExtractor
}

// newFlightReader creates a new flightReader.
func newFlightReader(stream flight.FlightService_DoGetClient, alloc memory.Allocator) (*flightReader, error) {
	extractor := &headerExtractor{stream: stream}
	reader, err := flight.NewRecordReader(extractor, ipc.WithAllocator(alloc))
	if err != nil {
		return nil, err
	}
	return &flightReader{
		Reader:    reader,
		extractor: extractor,
	}, nil
}

// Header returns the extracted headers.
func (s *flightReader) Header() (metadata.MD, error) {
	return s.extractor.Header()
}

// headerExtractor collects the stream's headers on the first call to Recv.
type headerExtractor struct {
	stream flight.FlightService_DoGetClient
	once   sync.Once
	header metadata.MD
	err    error
}

// Header returns the extracted headers.
func (s *headerExtractor) Header() (metadata.MD, error) {
	return s.header, s.err
}

// Recv reads from the stream and captures headers on the first call.
func (s *headerExtractor) Recv() (*flight.FlightData, error) {
	data, err := s.stream.Recv()
	s.once.Do(func() {
		s.header, s.err = s.stream.Header()
	})
	return data, err
}
