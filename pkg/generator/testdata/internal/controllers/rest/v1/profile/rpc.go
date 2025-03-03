package profile

import "github.com/sev-2/raiden"

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
	return `BEGIN INSERT INTO public.users (name, email, uid, avatar_url) VALUES ( NEW.raw_user_meta_data ->> 'name', NEW.raw_user_meta_data ->> 'email', NEW.id, NEW.raw_user_meta_data ->> 'avatar_url' ); RETURN NEW; END;`
}

type CreateProfileController struct {
	raiden.ControllerBase
	Payload *CreateProfileParams
	Result  CreateProfileResult
}

func (sc *CreateProfileController) Post(ctx raiden.Context) error {
	raiden.Info("[ScController] before get fire")
	return nil
}
