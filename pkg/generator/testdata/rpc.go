package testdata

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

type CreateProfileParams struct {
}
type CreateProfileResult interface{}

type CreateProfile struct {
	raiden.RpcBase
	Params *CreateProfileParams `json:"-"`
	Return CreateProfileResult  `json:"-"`
}

func (r *CreateProfile) GetName() string {
	return "create_profile"
}

func (r *CreateProfile) GetSecurity() raiden.RpcSecurityType {
	return raiden.RpcSecurityTypeDefiner
}

func (r *CreateProfile) GetReturnType() raiden.RpcReturnDataType {
	return raiden.RpcReturnDataTypeTrigger
}

func (r *CreateProfile) GetRawDefinition() string {
	return `BEGIN INSERT INTO public.users (firstname,lastname, email) VALUES ( NEW.raw_user_meta_data ->> 'name', NEW.raw_user_meta_data ->> 'name', NEW.raw_user_meta_data ->> 'email' ); RETURN NEW; END;`
}
