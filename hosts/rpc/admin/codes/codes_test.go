package codes

import (
	"testing"

	"github.com/jrapoport/gothic/core/codes"
	"github.com/jrapoport/gothic/core/context"
	"github.com/jrapoport/gothic/models/code"
	"github.com/jrapoport/gothic/test/tconn"
	"github.com/jrapoport/gothic/test/tsrv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodesServer_CreateSignupCodes(t *testing.T) {
	t.Parallel()
	const testLen = 10
	s, _ := tsrv.RPCServer(t, false)
	srv := newCodesServer(s)
	ctx := context.Background()
	_, err := srv.CreateSignupCodes(ctx, nil)
	assert.Error(t, err)
	req := &CreateSignupCodesRequest{
		Uses:  code.InfiniteUse,
		Count: testLen,
	}
	res, err := srv.CreateSignupCodes(ctx, req)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Len(t, res.GetCodes(), testLen)
}

func TestCodesServer_CheckSignupCode(t *testing.T) {
	t.Parallel()
	s, _ := tsrv.RPCServer(t, false)
	srv := newCodesServer(s)
	ctx := context.Background()
	cr, err := srv.CreateSignupCodes(ctx, &CreateSignupCodesRequest{
		Uses:  code.SingleUse,
		Count: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, cr)
	require.Len(t, cr.GetCodes(), 1)
	_, err = srv.CheckSignupCode(ctx, nil)
	assert.Error(t, err)
	req := &CheckSignupCodeRequest{
		Code: "bad",
	}
	_, err = srv.CheckSignupCode(ctx, req)
	assert.Error(t, err)
	test := cr.GetCodes()[0]
	req = &CheckSignupCodeRequest{
		Code: test,
	}
	res, err := srv.CheckSignupCode(ctx, req)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.True(t, res.Usable)
	assert.Equal(t, test, res.Code)
	conn := tconn.Conn(t, srv.Config())
	sc, err := codes.GetSignupCode(conn, test)
	require.NoError(t, err)
	require.NotNil(t, sc)
	sc.Used = 1
	err = conn.Save(sc).Error
	require.NoError(t, err)
	res, err = srv.CheckSignupCode(ctx, req)
	assert.NoError(t, err)
	assert.False(t, res.Usable)
	err = conn.Delete(sc).Error
	require.NoError(t, err)
	_, err = srv.CheckSignupCode(ctx, req)
	assert.Error(t, err)
}

func TestCodesServer_VoidSignupCode(t *testing.T) {
	t.Parallel()
	s, _ := tsrv.RPCServer(t, false)
	srv := newCodesServer(s)
	ctx := context.Background()
	_, err := srv.VoidSignupCode(ctx, nil)
	assert.Error(t, err)
	req := &VoidSignupCodeRequest{
		Code: "bad",
	}
	_, err = srv.VoidSignupCode(ctx, req)
	assert.Error(t, err)
	cr, err := srv.CreateSignupCodes(ctx, &CreateSignupCodesRequest{
		Uses:  code.SingleUse,
		Count: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, cr)
	require.Len(t, cr.GetCodes(), 1)
	test := cr.GetCodes()[0]
	req = &VoidSignupCodeRequest{
		Code: test,
	}
	_, err = srv.VoidSignupCode(ctx, req)
	assert.NoError(t, err)
	conn := tconn.Conn(t, srv.Config())
	_, err = codes.GetSignupCode(conn, test)
	require.Error(t, err)
}
