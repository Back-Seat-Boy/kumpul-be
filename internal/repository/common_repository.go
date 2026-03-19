package repository

import (
	"github.com/kumparan/cacher"
	"github.com/kumparan/go-utils"
	"github.com/kumparan/redsync/v4"
)

func findFromCacheByKey[T any](ck cacher.Keeper, key string) (res T, mu *redsync.Mutex, err error) {
	reply, mu, err := ck.GetOrLock(key)
	if err != nil || reply == nil {
		return
	}

	res = utils.InterfaceBytesToType[T](reply)
	return
}
