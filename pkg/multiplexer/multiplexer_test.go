package multiplexer

import (
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type MultiplexerSuite struct {
	suite.Suite
}

type selectorResult struct {
	selector Selector
	result   bool
}

func (s *MultiplexerSuite) TestOrSelector() {
	candidates := map[*http.Request]selectorResult{
		{
			ProtoMajor: 2,
			Header: map[string][]string{
				"Content-Type": {"application/grpc"},
			},
		}: {
			selector: OrSelector(
				IsGRPCRequest,
				IsPubSubRequest,
			),
			result: true,
		},
		{
			ProtoMajor: 2,
			Method:     http.MethodPost,
			Header: map[string][]string{
				"User-Agent": {"APIs-Google"},
			},
		}: {
			selector: OrSelector(
				IsGRPCRequest,
				IsPubSubRequest,
			),
			result: true,
		},
		{
			Header: map[string][]string{
				"Content-Type": {"application/json"},
			},
		}: {
			selector: OrSelector(
				IsGRPCRequest,
				IsPubSubRequest,
			),
			result: false,
		},
	}

	for req, res := range candidates {
		s.Equal(res.result, res.selector(req))
	}
}

func TestMultiplexerSuite(t *testing.T) {
	suite.Run(t, &MultiplexerSuite{})
}
