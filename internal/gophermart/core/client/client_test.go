package client

import (
	"context"
	"reflect"
	"testing"

	"gofermart/internal/gophermart/core/model"
)

func Test_handler_SendOrder(t *testing.T) {
	type fields struct {
		client     *req.Client
		ch         chan struct{}
		retryAfter *time.Duration
	}
	type args struct {
		ctx     context.Context
		orderID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    model.ClientResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &handler{
				client:     tt.fields.client,
				ch:         tt.fields.ch,
				retryAfter: tt.fields.retryAfter,
			}
			got, err := c.SendOrder(tt.args.ctx, tt.args.orderID)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SendOrder() got = %v, want %v", got, tt.want)
			}
		})
	}
}
