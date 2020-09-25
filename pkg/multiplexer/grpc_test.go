package multiplexer

import (
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type GRPCTestSuite struct {
	suite.Suite
}

func (s *GRPCTestSuite) TestIsGRPCRequest() {
	candidates := map[*http.Request]bool{
		&http.Request{
			Method:     http.MethodPost,
			ProtoMajor: 2,
			Header: map[string][]string{
				"Content-Type": {"application/grpc"},
			},
		}: true,
		&http.Request{
			Method: http.MethodPost,
			Header: map[string][]string{
				"Content-Type": {"application/json"},
			},
		}: false,
	}

	for req, result := range candidates {
		s.Equal(result, IsGRPCRequest(req), "Request is badly considered a grpc/non-grpc request", req.Header)
	}
}

func TestGRPCTestSuite(t *testing.T) {
	suite.Run(t, &GRPCTestSuite{})
}
