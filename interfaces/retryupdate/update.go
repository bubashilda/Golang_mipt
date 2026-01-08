//go:build !solution

package retryupdate

import (
	"errors"
	"github.com/gofrs/uuid"
	"gitlab.com/slon/shad-go/retryupdate/kvapi"
)

func UpdateValue(c kvapi.Client, key string, updateFn func(oldValue *string) (newValue string, err error)) error {
	var prevOldVersion = uuid.Nil
	for {
		getResp, err := c.Get(&kvapi.GetRequest{Key: key})
		var newValue string
		var errApi *kvapi.APIError

		if errors.Is(err, kvapi.ErrKeyNotFound) {
			newValue, err = updateFn(nil)
			if err != nil {
				return err
			}
		} else if errors.As(err, &errApi) {
			var errTemp *kvapi.AuthError
			if errors.As(errApi, &errTemp) {
				return err
			}
			continue
		} else if err != nil {
			continue
		} else {
			newValue, err = updateFn(&getResp.Value)
			if err != nil {
				return err
			}
		}

		for {
			var oldVersion uuid.UUID
			if getResp == nil {
				oldVersion = uuid.UUID{}
			} else {
				oldVersion = getResp.Version
			}

			newVersion := uuid.Must(uuid.NewV4())
			_, err = c.Set(&kvapi.SetRequest{
				Key:        key,
				Value:      newValue,
				OldVersion: oldVersion,
				NewVersion: newVersion},
			)

			var errConflict *kvapi.ConflictError
			var errApi *kvapi.APIError
			var errAuth *kvapi.AuthError

			if errors.As(err, &errApi) {
				if errors.Is(errApi, kvapi.ErrKeyNotFound) {
					newValue, err = updateFn(nil)
					if err != nil {
						return err
					}
					getResp = nil
				} else if errors.As(errApi, &errAuth) {
					return err
				} else if errors.As(errApi, &errConflict) {
					if errConflict.ExpectedVersion == prevOldVersion {
						return nil
					}
					break
				}
			} else if err == nil {
				return nil
			}
			prevOldVersion = newVersion
		}
	}
}
