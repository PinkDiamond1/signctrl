package rpc

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	tm_json "github.com/tendermint/tendermint/libs/json"
	tm_coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tm_types "github.com/tendermint/tendermint/types"
)

// getFreePort asks the kernel for a free port that is ready to use.
func getFreePort(t *testing.T) (port int, err error) {
	t.Helper()
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}

	return
}

func testBlockResult(t *testing.T) *BlockResult {
	t.Helper()
	return &BlockResult{
		jsonrpc: "2.0",
		id:      1,
		Result: &tm_coretypes.ResultBlock{
			Block: &tm_types.Block{
				LastCommit: &tm_types.Commit{
					Signatures: []tm_types.CommitSig{
						{
							ValidatorAddress: []byte("ALPHA-ADDR"),
							Signature:        []byte("ALPHA-SIG"),
						},
						{
							ValidatorAddress: []byte("BETA-ADDR"),
							Signature:        []byte("BETA-SIG"),
						},
					},
				},
			},
		},
	}
}

func TestQueryBlock(t *testing.T) {
	// Invalid block height.
	port, _ := getFreePort(t)
	addr := fmt.Sprintf("tcp://127.0.0.1:%v", port)
	rb, err := QueryBlock(context.Background(), addr, 1, log.New(ioutil.Discard, "", 0))
	assert.Nil(t, rb)
	assert.Error(t, err)

	// Valid block height.
	port, _ = getFreePort(t)
	addr = fmt.Sprintf("tcp://127.0.0.1:%v", port)
	go func() {
		http.HandleFunc("/block", func(rw http.ResponseWriter, r *http.Request) {
			height := r.URL.Query().Get("height")
			assert.Equal(t, "2", height)

			bytes, _ := tm_json.Marshal(testBlockResult(t))
			rw.Write(bytes)
		})
		http.ListenAndServe(strings.TrimPrefix(addr, "tcp://"), nil)
	}()

	rb, err = QueryBlock(context.Background(), addr, 2, log.New(ioutil.Discard, "", 0))
	assert.NotNil(t, rb)
	assert.NoError(t, err)
}
