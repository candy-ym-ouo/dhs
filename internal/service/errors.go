package service

import "errors"

func flattenError(err error) error {
	if err == nil {
		return nil
	}
	return errors.New(err.Error())
}
