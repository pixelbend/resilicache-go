package zcnatskv

import (
	"context"
	"encoding/json"
	"github.com/driftdev/zencache"
	"github.com/driftdev/zencache/zcerror"
	"github.com/nats-io/nats.go"
	"time"
)

type Backend struct {
	client nats.KeyValue
}

var _ zencache.IZenCache = (*Backend)(nil)

func NewBackend(client nats.KeyValue) *Backend {
	return &Backend{
		client: client,
	}
}

func (b *Backend) Set(ctx context.Context, key string, data any, expiry time.Duration) error {
	item := zencache.NewItem(data)
	item.SetExpiration(expiry)

	itemBytes, err := json.Marshal(item)
	if err != nil {
		return err
	}

	_, err = b.client.Put(key, itemBytes)
	if err != nil {
		return err
	}

	return nil
}

func (b *Backend) Get(ctx context.Context, key string, data any) error {
	result, err := b.client.Get(key)
	if err != nil {
		return err
	}
	if result == nil {
		return zcerror.ErrorValueNotFound
	}

	var item zencache.Item
	err = json.Unmarshal(result.Value(), &item)
	if err != nil {
		return err
	}

	if item.IsExpired() {
		err := b.Delete(ctx, key)
		if err != nil {
			return err
		}
		return zcerror.ErrorValueNotFound
	}

	err = item.ParseData(&data)
	if err != nil {
		return err
	}

	return nil
}

func (b *Backend) Delete(ctx context.Context, key string) error {
	err := b.client.Purge(key)
	if err != nil {
		return err
	}

	return nil
}