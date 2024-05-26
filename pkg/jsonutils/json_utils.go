package jsonutils

import (
	"encoding/json"

	"github.com/NikitaTsaralov/utils/logger"
	"github.com/pkg/errors"
)

func UnsafeMarshall(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		logger.Error(errors.Wrapf(err, "UnsafeMarshall()"))
	}

	return data
}
