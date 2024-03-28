package admin

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/supabase/client"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/valyala/fasthttp"
)

func UpdateUser(cfg *raiden.Config, userId string, payload objects.User) (user objects.User, err error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return user, err
	}

	basePublicUrl := strings.TrimRight(cfg.SupabasePublicUrl, "/")
	url := fmt.Sprintf("%s/auth/v1/admin/users/%s", basePublicUrl, userId)

	auth := func(req *fasthttp.Request) error {
		req.Header.Set("apikey", cfg.ServiceKey)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.ServiceKey))
		return nil
	}
	return client.Put[objects.User](url, body, client.DefaultTimeout, auth, nil)
}
