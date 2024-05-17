package supabase

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/sev-2/raiden/pkg/logger"
	"github.com/sev-2/raiden/pkg/supabase/client/net"
	"github.com/sev-2/raiden/pkg/supabase/objects"
)

var StorageLogger = logger.HcLog().Named("supabase.storage")

type CreateBucketSuccessResponse struct {
	Name string `json:"name"`
}

type DefaultBucketSuccessResponse struct {
	Message string `json:"message"`
}

func DefaultAuthInterceptor(apiKey string, accessToken string) func(req *http.Request) error {
	return func(req *http.Request) error {
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
	return decorateActionWithDataErr("fetch", "storage", func() ([]objects.Bucket, error) {
		StorageLogger.Debug("fetch all bucket")
		return net.Get[[]objects.Bucket](getBucketUrl(cfg), net.DefaultTimeout, DefaultAuthInterceptor(cfg.ServiceKey, cfg.ServiceKey), nil)
	})
}

func GetBucket(cfg *raiden.Config, name string) (buckets objects.Bucket, err error) {
	return decorateActionWithDataErr("fetch", "storage", func() (objects.Bucket, error) {
		StorageLogger.Debug("fetch bucket")
		url := fmt.Sprintf("%s/%s", getBucketUrl(cfg), name)
		return net.Get[objects.Bucket](url, net.DefaultTimeout, DefaultAuthInterceptor(cfg.ServiceKey, cfg.ServiceKey), nil)
	})

}

func CreateBucket(cfg *raiden.Config, param objects.Bucket) (bucket objects.Bucket, err error) {
	return decorateActionWithDataErr("create", "storage", func() (objects.Bucket, error) {
		StorageLogger.Debug("start create storage", "name", param.Name)
		byteData, err := json.Marshal(param)
		if err != nil {
			return bucket, err
		}

		res, err := net.Post[CreateBucketSuccessResponse](
			getBucketUrl(cfg), byteData, net.DefaultTimeout, DefaultAuthInterceptor(cfg.ServiceKey, cfg.ServiceKey), nil,
		)

		if err != nil {
			return bucket, err
		}
		StorageLogger.Debug("finish create storage", "name", param.Name)
		return GetBucket(cfg, res.Name)
	})
}

func UpdateBucket(cfg *raiden.Config, param objects.Bucket, updateItem objects.UpdateBucketParam) error {
	return decorateActionErr("update", "storage", func() error {
		StorageLogger.Debug("start update storage", "name", param.Name)

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
		_, err = net.Put[DefaultBucketSuccessResponse](
			url, byteData, net.DefaultTimeout, DefaultAuthInterceptor(cfg.ServiceKey, cfg.ServiceKey), nil,
		)
		StorageLogger.Debug("finish update storage", "name", param.Name)
		return err
	})
}

func DeleteBucket(cfg *raiden.Config, param objects.Bucket) (err error) {
	return decorateActionErr("delete", "storage", func() error {
		StorageLogger.Debug("start delete storage", "name", param.Name)

		deleteReqInterceptor := func(req *http.Request) error {
			req.Header.Set("apiKey", cfg.ServiceKey)
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cfg.ServiceKey))
			req.Header.Set("Content-Type", "")
			return nil
		}

		url := fmt.Sprintf("%s/%s", getBucketUrl(cfg), param.ID)
		_, err = net.Delete[DefaultBucketSuccessResponse](
			url, nil, net.DefaultTimeout, deleteReqInterceptor, nil,
		)

		StorageLogger.Debug("finish delete storage", "name", param.Name)
		return err
	})
}
