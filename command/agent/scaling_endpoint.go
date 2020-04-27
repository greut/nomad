package agent

import (
	"net/http"
	"strings"

	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/nomad/sdk/structs"
)

func (s *HTTPServer) ScalingPoliciesRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	switch req.Method {
	case "GET":
		return s.scalingPoliciesListRequest(resp, req)
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

func (s *HTTPServer) scalingPoliciesListRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	args := structs.ScalingPolicyListRequest{}
	if s.parse(resp, req, &args.Region, &args.QueryOptions) {
		return nil, nil
	}

	var out structs.ScalingPolicyListResponse
	if err := s.agent.RPC("Scaling.ListPolicies", &args, &out); err != nil {
		return nil, err
	}

	setMeta(resp, &out.QueryMeta)
	if out.Policies == nil {
		out.Policies = make([]*structs.ScalingPolicyListStub, 0)
	}
	return out.Policies, nil
}

func (s *HTTPServer) ScalingPolicySpecificRequest(resp http.ResponseWriter, req *http.Request) (interface{}, error) {
	path := strings.TrimPrefix(req.URL.Path, "/v1/scaling/policy/")
	switch {
	default:
		return s.scalingPolicyCRUD(resp, req, path)
	}
}

func (s *HTTPServer) scalingPolicyCRUD(resp http.ResponseWriter, req *http.Request,
	policyID string) (interface{}, error) {
	switch req.Method {
	case "GET":
		return s.scalingPolicyQuery(resp, req, policyID)
	default:
		return nil, CodedError(405, ErrInvalidMethod)
	}
}

func (s *HTTPServer) scalingPolicyQuery(resp http.ResponseWriter, req *http.Request,
	policyID string) (interface{}, error) {
	args := structs.ScalingPolicySpecificRequest{
		ID: policyID,
	}
	if s.parse(resp, req, &args.Region, &args.QueryOptions) {
		return nil, nil
	}

	var out structs.SingleScalingPolicyResponse
	if err := s.agent.RPC("Scaling.GetPolicy", &args, &out); err != nil {
		return nil, err
	}

	setMeta(resp, &out.QueryMeta)
	if out.Policy == nil {
		return nil, CodedError(404, "policy not found")
	}

	return out.Policy, nil
}

func ApiScalingPolicyToStructs(count int, ap *api.ScalingPolicy) *structs.ScalingPolicy {
	p := structs.ScalingPolicy{
		Enabled: *ap.Enabled,
		Max:     ap.Max,
		Policy:  ap.Policy,
		Target:  map[string]string{},
	}
	if ap.Min != nil {
		p.Min = *ap.Min
	} else {
		p.Min = int64(count)
	}
	return &p
}
