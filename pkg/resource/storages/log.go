package storages

import (
	"github.com/hashicorp/go-hclog"
	"github.com/sev-2/raiden/pkg/logger"
)

var Logger hclog.Logger = logger.HcLog().Named("resource.storages")
