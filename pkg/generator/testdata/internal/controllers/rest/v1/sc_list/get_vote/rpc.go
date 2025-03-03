package sclist

import "github.com/sev-2/raiden"

type GetVoteByParams struct {
	CandidateName string `json:"candidate_name" column:"name:candidate_name;type:varchar;default:anon"`
	VoterName     string `json:"voter_name" column:"name:voter_name;type:varchar;default:anon"`
}
type GetVoteByResult any

type GetVoteBy struct {
	raiden.RpcBase
	Params *GetVoteByParams `json:"-"`
	Return GetVoteByResult  `json:"-"`
}

type GetVoteController struct {
	raiden.ControllerBase
	Payload *GetVoteByParams
	Result  GetVoteByResult
}

func (sc *GetVoteController) Post(ctx raiden.Context) error {
	raiden.Info("[ScController] before get fire")
	return nil
}
