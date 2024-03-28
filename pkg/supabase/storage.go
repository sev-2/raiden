package supabase

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/client"
	"github.com/sev-2/raiden/pkg/supabase/objects"
	"github.com/valyala/fasthttp"
)

type CreateBucketSuccessResponse struct {
	Name string `json:"name"`
}

type DefaultBucketSuccessResponse struct {
	Message string `json:"message"`
}

func DefaultAuthInterceptor(apiKey string, accessToken string) func(req *fasthttp.Request) error {
	return func(req *fasthttp.Request) error {
		req.Header.Set("apiKey", apiKey)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
		return nil
	}
}

func getBucketUrl(cfg *raiden.Config) string {
	publicUrl := strings.TrimSuffix(cfg.SupabasePublicUrl, "/")
	return fmt.Sprintf("%s/storage/v1/bucket", publicUrl)
}

func GetBuckets(cfg *raiden.Config) (buckets []objects.Bucket, err error) {
	logger.Debug("get all bucket from storage")
	return client.Get[[]objects.Bucket](
		getBucketUrl(cfg), client.DefaultTimeout, DefaultAuthInterceptor(cfg.ServiceKey, cfg.ServiceKey), nil,
	)
}

func GetBucket(cfg *raiden.Config, bucketId string) (buckets objects.Bucket, err error) {
	logger.Debug("get bucket from storage")
	url := fmt.Sprintf("%s/%s", getBucketUrl(cfg), bucketId)
	return client.Get[objects.Bucket](url, client.DefaultTimeout, DefaultAuthInterceptor(cfg.ServiceKey, cfg.ServiceKey), nil)
}

func CreateBucket(cfg *raiden.Config, param objects.Bucket) (bucket objects.Bucket, err error) {
	logger.Debug("create new bucket")
	byteData, err := json.Marshal(param)
	if err != nil {
		return bucket, err
	}

	logger.PrintJson(param, true)
	res, err := client.Post[CreateBucketSuccessResponse](
		getBucketUrl(cfg), byteData, client.DefaultTimeout, DefaultAuthInterceptor(cfg.ServiceKey, cfg.ServiceKey), nil,
	)

	if err != nil {
		return bucket, err
	}

	return GetBucket(cfg, res.Name)
}

func UpdateBucket(cfg *raiden.Config, param objects.Bucket, updateItem objects.UpdateBucketParam) error {
	logger.Debug("update bucket")

	// just return, nothing to be processed
	if len(updateItem.ChangeItems) == 0 {
		return nil
	}

	// build update payload
	updateBucket := objects.Bucket{}
	for i := range updateItem.ChangeItems {
		u := updateItem.ChangeItems[i]
		switch u {
		case objects.UpdateBucketIsPublic:
			updateBucket.Public = param.Public
		case objects.UpdateBucketFileSizeLimit:
			updateBucket.FileSizeLimit = param.FileSizeLimit
		case objects.UpdateBucketAllowedMimeTypes:
			updateBucket.AllowedMimeTypes = param.AllowedMimeTypes
		}
	}

	// build request
	byteData, err := json.Marshal(updateBucket)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/%s", getBucketUrl(cfg), param.ID)
	_, err = client.Put[DefaultBucketSuccessResponse](
		url, byteData, client.DefaultTimeout, DefaultAuthInterceptor(cfg.ServiceKey, cfg.ServiceKey), nil,
	)
	return err
}

func DeleteBucket(cfg *raiden.Config, param objects.Bucket) (err error) {
	logger.Debug("delete bucket")

	deleteReqInterceptor := func(req *fasthttp.Request) error {
		req.Header.Set("apiKey", cfg.ServiceKey)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.ServiceKey))
		req.Header.SetContentType("")
		return nil
	}

	url := fmt.Sprintf("%s/%s", getBucketUrl(cfg), param.ID)
	_, err = client.Delete[DefaultBucketSuccessResponse](
		url, nil, client.DefaultTimeout, deleteReqInterceptor, nil,
	)
	return err
}
